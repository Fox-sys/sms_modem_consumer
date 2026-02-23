package domain

type SmsReader interface {
	Connect() error
	GetSMS(opts GetSMSOpts) ([]SmsMessage, error)
}

type GetSMSOpts struct {
	PageIndex       int
	ReadCount       int
	BoxType         int
	DeleteAfterRead bool
}

type SmsForwarder interface {
	Forward(messages []SmsMessage) error
}
