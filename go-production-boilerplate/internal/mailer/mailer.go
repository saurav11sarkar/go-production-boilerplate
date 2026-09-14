package mailer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/config"
)

type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

type Mailer interface {
	Send(ctx context.Context, message Message) error
}

func New(cfg config.Config, logger *slog.Logger) (Mailer, error) {
	switch cfg.Mail.Provider {
	case "log":
		return &LogMailer{logger: logger}, nil
	case "smtp":
		if cfg.Mail.Host == "" || cfg.Mail.From == "" {
			return nil, fmt.Errorf("SMTP_HOST and MAIL_FROM are required when MAIL_PROVIDER=smtp")
		}
		return &SMTPMailer{
			fromName: cfg.Mail.FromName,
			from:     cfg.Mail.From,
			host:     cfg.Mail.Host,
			port:     cfg.Mail.Port,
			username: cfg.Mail.Username,
			password: cfg.Mail.Password,
			timeout:  cfg.Mail.Timeout,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported mail provider %q", cfg.Mail.Provider)
	}
}
