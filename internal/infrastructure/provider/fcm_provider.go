package provider

import (
	"log/slog"

	"gorm.io/gorm"
	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
)

type FCMProvider struct {
	logger *slog.Logger
	apiKey string
	db     *gorm.DB
}

func NewFCMProvider(logger *slog.Logger, apiKey string, db *gorm.DB) *FCMProvider {
	return &FCMProvider{
		logger: logger,
		apiKey: apiKey,
		db:     db,
	}
}

func (p *FCMProvider) Send(recipient string, eventType string, payload map[string]interface{}) error {
	p.logger.Info("Sending Push Notification via FCM", "recipient", recipient, "event_type", eventType)
	
	// Assuming recipient here is the userID because notification_usecase routes it
	// Actually we should fetch user_id. Let's just fetch devices.
	userID := recipient
	if uid, ok := payload["user_id"].(string); ok {
		userID = uid
	}

	var devices []domain.UserDeviceModel
	if err := p.db.Where("user_id = ?", userID).Find(&devices).Error; err != nil {
		p.logger.Error("Failed to fetch devices for FCM", "error", err)
		return err
	}

	for _, device := range devices {
		p.logger.Info("Dispatching FCM", "token", device.DeviceToken)
		// actual FCM call using device.DeviceToken
	}

	return nil
}
