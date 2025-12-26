package email

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)

var (
	ErrMissingRecipient = errors.New("missing email recipient")
	ErrMissingSubject   = errors.New("missing email subject")
	ErrMissingBody      = errors.New("missing email body")
	ErrMissingSender    = errors.New("missing email sender")
	ErrMissingSMTPHost  = errors.New("missing SMTP host")
)

// Message represents an email message.
type Message struct {
	From        string
	To          []string
	Subject     string
	Body        string
	ContentType string
}

// Sender defines email sending behaviour.
type Sender interface {
	Send(ctx context.Context, message Message) error
}

// SMTPConfig configures an SMTP sender.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// SMTPSender sends email via SMTP.
type SMTPSender struct {
	host string
	port int
	from string
	auth smtp.Auth
}

// NoopSender discards emails (useful in development).
type NoopSender struct{}

// NewNoopSender returns a no-op sender.
func NewNoopSender() Sender {
	return NoopSender{}
}

// Send discards the email message.
func (NoopSender) Send(ctx context.Context, message Message) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

// NewSMTPSender creates a new SMTP sender.
func NewSMTPSender(cfg SMTPConfig) (*SMTPSender, error) {
	if cfg.Host == "" {
		return nil, ErrMissingSMTPHost
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.From == "" {
		return nil, ErrMissingSender
	}

	var auth smtp.Auth
	if cfg.Username != "" || cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return &SMTPSender{
		host: cfg.Host,
		port: cfg.Port,
		from: cfg.From,
		auth: auth,
	}, nil
}

// Send delivers an email message.
func (s *SMTPSender) Send(ctx context.Context, message Message) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	if len(message.To) == 0 {
		return ErrMissingRecipient
	}
	if strings.TrimSpace(message.Subject) == "" {
		return ErrMissingSubject
	}
	if strings.TrimSpace(message.Body) == "" {
		return ErrMissingBody
	}

	from := strings.TrimSpace(message.From)
	if from == "" {
		from = s.from
	}
	if from == "" {
		return ErrMissingSender
	}

	contentType := message.ContentType
	if contentType == "" {
		contentType = "text/plain; charset=\"utf-8\""
	}

	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", strings.Join(message.To, ", ")),
		fmt.Sprintf("Subject: %s", sanitizeHeader(message.Subject)),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: %s", contentType),
		"",
	}
	payload := strings.Join(headers, "\r\n") + message.Body

	return smtp.SendMail(s.addr(), s.auth, from, message.To, []byte(payload))
}

func (s *SMTPSender) addr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

func sanitizeHeader(value string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(value)
}
