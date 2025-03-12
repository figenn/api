package mailer

import (
	"context"
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

type Config struct {
	To      string
	Html    string
	Subject string
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
	if len(config.To) == 0 || config.Html == "" || config.Subject == "" {
		return "", fmt.Errorf("from, to, html and subject fields are required")
	}

	params := &resend.SendEmailRequest{
		From:    os.Getenv("SENDER_EMAIL"),
		To:      []string{config.To},
		Html:    config.Html,
		Subject: config.Subject,
	}

	sent, err := m.client.Emails.Send(params)
	if err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return sent.Id, nil
}
