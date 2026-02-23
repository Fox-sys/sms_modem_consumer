package huawei

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"sender-modem/src/internal/domain"
)

var ErrNotConnected = errors.New("adapter not connected: call Connect first")

const defaultHTTPTimeout = 30 * time.Second

type Adapter struct {
	baseURL   string
	username  string
	password  string
	client    *http.Client
	token     string
	sessionID string
}

func NewAdapter(baseURL, username, password string) *Adapter {
	baseURL = strings.TrimSuffix(baseURL, "/")
	jar, _ := cookiejar.New(nil)
	return &Adapter{
		baseURL:  baseURL,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: defaultHTTPTimeout,
			Jar:     jar,
		},
	}
}

func (a *Adapter) Connect() error {
	tok, sess, err := a.getSessionToken()
	if err != nil {
		return err
	}
	loginTok, newSess, err := a.login(tok, sess)
	if err != nil {
		return err
	}
	a.token = loginTok
	if newSess != "" {
		a.sessionID = newSess
	} else {
		a.sessionID = sess
	}
	return nil
}

func (a *Adapter) GetSMS(opts domain.GetSMSOpts) ([]domain.SmsMessage, error) {
	if a.token == "" || a.sessionID == "" {
		return nil, ErrNotConnected
	}
	if opts.PageIndex == 0 {
		opts.PageIndex = 1
	}
	if opts.ReadCount == 0 {
		opts.ReadCount = 50
	}
	if opts.BoxType == 0 {
		opts.BoxType = 1
	}

	body := fmt.Sprintf(`<request>
<PageIndex>%d</PageIndex>
<ReadCount>%d</ReadCount>
<BoxType>%d</BoxType>
<SortType>0</SortType>
<Ascending>0</Ascending>
<UnreadPreferred>0</UnreadPreferred>
</request>`, opts.PageIndex, opts.ReadCount, opts.BoxType)

	req, err := http.NewRequest(http.MethodPost, a.baseURL+"/api/sms/sms-list", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	a.setAuth(req)
	req.Header.Set("Content-Type", "application/xml")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if nextTok := resp.Header.Get("__RequestVerificationToken"); nextTok != "" {
		a.token = nextTok
	}
	xmlStr := string(data)
	if err := a.checkError(xmlStr); err != nil {
		return nil, err
	}

	messages := a.parseSMS(xmlStr)
	if opts.DeleteAfterRead && len(messages) > 0 {
		indices := make([]int, len(messages))
		for i := range messages {
			indices[i] = messages[i].Index
		}
		if err := a.deleteSMS(indices); err != nil {
			return messages, err
		}
	}
	return messages, nil
}

func (a *Adapter) setAuth(req *http.Request) {
	req.Header.Set("__RequestVerificationToken", a.token)
	cookieVal := a.sessionID
	if cookieVal != "" && !strings.HasPrefix(cookieVal, "SessionID=") {
		cookieVal = "SessionID=" + cookieVal
	}
	req.Header.Set("Cookie", cookieVal)
}

func (a *Adapter) getSessionToken() (token, session string, err error) {
	resp, err := a.client.Get(a.baseURL + "/api/webserver/SesTokInfo")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	data = bytes.TrimSpace(data)
	var info struct {
		TokInfo string `xml:"TokInfo"`
		SesInfo string `xml:"SesInfo"`
	}
	if err := xml.Unmarshal(data, &info); err != nil {
		return "", "", err
	}
	info.TokInfo = strings.TrimSpace(info.TokInfo)
	info.SesInfo = strings.TrimSpace(info.SesInfo)
	if info.TokInfo == "" || info.SesInfo == "" {
		return "", "", errors.New("failed to obtain session/token")
	}
	return info.TokInfo, info.SesInfo, nil
}

func (a *Adapter) login(token, sessionID string) (loginToken, newSessionID string, err error) {
	body := fmt.Sprintf(`<request>
<Username>%s</Username>
<Password>%s</Password>
<password_type>4</password_type>
</request>`, a.username, a.password)

	req, err := http.NewRequest(http.MethodPost, a.baseURL+"/api/user/login", strings.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("__RequestVerificationToken", token)
	cookieVal := sessionID
	if cookieVal != "" && !strings.HasPrefix(cookieVal, "SessionID=") {
		cookieVal = "SessionID=" + cookieVal
	}
	req.Header.Set("Cookie", cookieVal)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if !bytes.Contains(data, []byte("<response>OK</response>")) {
		return "", "", errors.New("login failed")
	}
	loginToken = resp.Header.Get("__RequestVerificationToken")
	if loginToken != "" && strings.Contains(loginToken, "#") {
		loginToken = strings.Split(loginToken, "#")[0]
	}
	if loginToken == "" {
		return "", "", errors.New("no verification token returned after login")
	}
	for _, c := range resp.Cookies() {
		if c.Name == "SessionID" {
			newSessionID = c.Value
			break
		}
	}
	return loginToken, newSessionID, nil
}

func (a *Adapter) deleteSMS(indices []int) error {
	var b strings.Builder
	for _, i := range indices {
		b.WriteString(fmt.Sprintf("<Index>%d</Index>", i))
	}
	body := "<request>\n" + b.String() + "\n</request>"

	req, err := http.NewRequest(http.MethodPost, a.baseURL+"/api/sms/delete-sms", strings.NewReader(body))
	if err != nil {
		return err
	}
	a.setAuth(req)
	req.Header.Set("Content-Type", "application/xml")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if nextTok := resp.Header.Get("__RequestVerificationToken"); nextTok != "" {
		a.token = nextTok
	}
	if err := a.checkError(string(data)); err != nil {
		return err
	}
	if !bytes.Contains(data, []byte("<response>OK</response>")) {
		return errors.New("failed to delete SMS")
	}
	return nil
}

func (a *Adapter) checkError(xmlStr string) error {
	if !strings.Contains(xmlStr, "<error>") {
		return nil
	}
	var errResp struct {
		XMLName xml.Name `xml:"error"`
		Code    string   `xml:"code"`
	}
	if err := xml.Unmarshal([]byte(xmlStr), &errResp); err != nil {
		return fmt.Errorf("modem API error")
	}
	return fmt.Errorf("modem API error: %s", strings.TrimSpace(errResp.Code))
}

func (a *Adapter) parseSMS(xmlStr string) []domain.SmsMessage {
	type msgStruct struct {
		Index   int    `xml:"Index"`
		Phone   string `xml:"Phone"`
		Content string `xml:"Content"`
		Date    string `xml:"Date"`
		Smstat  int    `xml:"Smstat"`
		SmsType int    `xml:"SmsType"`
	}
	var root struct {
		XMLName  xml.Name   `xml:"response"`
		Messages []msgStruct `xml:"Messages>Message"`
	}
	if err := xml.Unmarshal([]byte(xmlStr), &root); err != nil {
		return nil
	}
	out := make([]domain.SmsMessage, 0, len(root.Messages))
	for _, m := range root.Messages {
		out = append(out, domain.SmsMessage{
			Index:   m.Index,
			Phone:   strings.TrimSpace(m.Phone),
			Content: html.UnescapeString(m.Content),
			Date:    strings.TrimSpace(m.Date),
			Smstat:  m.Smstat,
			SmsType: m.SmsType,
		})
	}
	return out
}
