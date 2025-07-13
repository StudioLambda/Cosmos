package contract

import "context"

type Mail struct {
	Subject     string
	To          []string
	From        string
	CC          []string
	BCC         []string
	Body        []byte
	ContentType string
}

type Mailer interface {
	Send(context context.Context, mail Mail) error
}
