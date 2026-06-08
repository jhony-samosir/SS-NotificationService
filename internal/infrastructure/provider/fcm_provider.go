package provider

import (
	"log/slog"
)

type FCMProvider struct {
	logger *slog.Logger
	apiKey string
}

func NewFCMProvider(logger *slog.Logger, apiKey string) *FCMProvider {
	return &FCMProvider{
		logger: logger,
		apiKey: apiKey,
	}
}

func (p *FCMProvider) Send(recipient string, eventType string, payload map[string]interface{}) error {
	p.logger.Info("Sending Push Notification via FCM", "recipient", recipient, "event_type", eventType)
	// Placeholder for actual Firebase Cloud Messaging API call
	return nil
}
