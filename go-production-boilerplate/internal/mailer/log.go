package mailer

import (
	"context"
	"log/slog"
)

type LogMailer struct {
	logger *slog.Logger
}

func (m *LogMailer) Send(ctx context.Context, message Message) error {
	m.logger.InfoContext(ctx, "development email",
		"to", message.To,
		"subject", message.Subject,
		"text", message.Text,
	)
	return nil
}
