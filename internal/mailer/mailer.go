package mailer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/resend/resend-go/v2"
)

type Config struct {
	To      string
	Html    string
	Subject string
}
type Mailer interface {
	SendMail(ctx context.Context, config Config) (string, error)
}

type resendMailer struct {
	client *resend.Client
}

type mailhogMailer struct {
	host     string
	port     string
	from     string
	username string
	password string
}

func NewMailer() Mailer {
	env := os.Getenv("APP_ENV")

	if env == "local" {
		return NewMailhogMailer(
			os.Getenv("SMTP_HOST"),
			os.Getenv("SMTP_PORT"),
			os.Getenv("SMTP_FROM"),
			os.Getenv("SMTP_USERNAME"),
			os.Getenv("SMTP_PASSWORD"),
		)
	}

	return NewResendMailer(os.Getenv("RESEND_API_KEY"))
}

func NewResendMailer(apiKey string) Mailer {
	return &resendMailer{
		client: resend.NewClient(apiKey),
	}
}

func NewMailhogMailer(host, port, from, username, password string) Mailer {
	if host == "" || port == "" || from == "" {
		log.Fatal("SMTP_HOST, SMTP_PORT and SMTP_FROM are required")
		return nil
	}

	return &mailhogMailer{
		host:     host,
		port:     port,
		from:     from,
		username: username,
		password: password,
	}
}

func (m *resendMailer) SendMail(ctx context.Context, config Config) (string, error) {
	if len(config.To) == 0 || config.Html == "" || config.Subject == "" {
		return "", fmt.Errorf("to, html and subject fields are required")
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

func (m *mailhogMailer) SendMail(ctx context.Context, config Config) (string, error) {
	if len(config.To) == 0 || config.Html == "" || config.Subject == "" {
		return "", errors.New("to, html and subject fields are required")
	}

	headers := make(map[string]string)
	headers["From"] = m.from
	headers["To"] = config.To
	headers["Subject"] = config.Subject
	headers["MIME-version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + config.Html

	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	var err error
	if m.username == "" && m.password == "" {
		err = smtp.SendMail(
			addr,
			nil,
			m.from,
			[]string{config.To},
			[]byte(message),
		)
	} else {
		err = smtp.SendMail(
			addr,
			auth,
			m.from,
			[]string{config.To},
			[]byte(message),
		)
	}

	if err != nil {
		return "", errors.New("failed to send email")
	}

	id := fmt.Sprintf("dev_%d", ctx.Value("request_id"))

	return id, nil
}
