package provider

import (
	"log/slog"

	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
)

type TwilioProvider struct {
	logger *slog.Logger
	apiKey string
}

func NewTwilioProvider(logger *slog.Logger, apiKey string) *TwilioProvider {
	return &TwilioProvider{
		logger: logger,
		apiKey: apiKey,
	}
}

func (p *TwilioProvider) Send(recipient string, eventType string, payload map[string]interface{}) error {
	p.logger.Info("Sending SMS via Twilio", "recipient", recipient, "event_type", eventType)
	// Placeholder for actual Twilio API call
	return nil
}
