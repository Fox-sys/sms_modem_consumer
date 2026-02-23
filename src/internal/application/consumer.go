package application

import (
	"log/slog"
	"time"

	"sender-modem/src/internal/domain"
)

type PollAndForward struct {
	Reader    domain.SmsReader
	Forwarder domain.SmsForwarder
	Interval  time.Duration
}

func (c *PollAndForward) Run() error {
	if err := c.Reader.Connect(); err != nil {
		return err
	}
	slog.Info("connected to modem", "interval_sec", int(c.Interval.Seconds()))
	for {
		messages, err := c.Reader.GetSMS(domain.GetSMSOpts{
			PageIndex:       1,
			ReadCount:       50,
			BoxType:         1,
			DeleteAfterRead: true,
		})
		if err != nil {
			slog.Error("poll cycle failed", "err", err)
			time.Sleep(c.Interval)
			continue
		}
		if len(messages) > 0 {
			slog.Info("forwarding messages", "count", len(messages))
			for _, m := range messages {
				slog.Debug("message", "phone", m.Phone, "content", m.Content, "date", m.Date)
			}
			if err := c.Forwarder.Forward(messages); err != nil {
				slog.Error("forward failed", "err", err)
			}
		}
		time.Sleep(c.Interval)
	}
}
