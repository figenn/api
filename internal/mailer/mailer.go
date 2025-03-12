package mailer

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

type Config struct {
	From    string
	To      string
	Html    string
	Subject string
	Cc      []string
	Bcc     []string
	ReplyTo string
}

// mockgen -source=internal/mailer/mailer.go -destination=internal/mailer/mocks/mock_mailer.go -package=mocks
type Mailer interface {
	SendMail(ctx context.Context, config Config) (string, error)
}

type resendMailer struct {
	client *resend.Client
}

func NewMailer(apiKey string) Mailer {
	return &resendMailer{
		client: resend.NewClient(apiKey),
	}
}

func (m *resendMailer) SendMail(ctx context.Context, config Config) (string, error) {
	if config.From == "" || len(config.To) == 0 || config.Html == "" || config.Subject == "" {
		return "", fmt.Errorf("from, to, html and subject fields are required")
	}

	params := &resend.SendEmailRequest{
		From:    config.From,
		To:      []string{config.To},
		Html:    config.Html,
		Subject: config.Subject,
		Cc:      config.Cc,
		Bcc:     config.Bcc,
		ReplyTo: config.ReplyTo,
	}

	sent, err := m.client.Emails.Send(params)
	if err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return sent.Id, nil
}
