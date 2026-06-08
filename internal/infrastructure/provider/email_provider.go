package provider

import (
	"log/slog"
	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
)

type SendGridProvider struct {
	logger *slog.Logger
	apiKey string
}

func NewSendGridProvider(logger *slog.Logger, apiKey string) domain.NotificationProvider {
	return &SendGridProvider{
		logger: logger,
		apiKey: apiKey,
	}
}

func (p *SendGridProvider) Send(recipient string, templateName string, data map[string]interface{}) error {
	p.logger.Info("Mock sending email via SendGrid", slog.String("recipient", recipient), slog.String("template", templateName))
	// In reality: build sendgrid client, construct email, send.
	return nil
}
