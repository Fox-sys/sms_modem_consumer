package domain

type SmsMessage struct {
	Index   int    `json:"index"`
	Phone   string `json:"phone"`
	Content string `json:"content"`
	Date    string `json:"date"`
	Smstat  int    `json:"smstat"`
	SmsType int    `json:"sms_type"`
}
