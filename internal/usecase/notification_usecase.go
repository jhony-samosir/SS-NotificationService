package usecase

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
)

type NotificationUsecase struct {
	emailProvider domain.NotificationProvider
	logger        *slog.Logger
}

func NewNotificationUsecase(
	emailProvider domain.NotificationProvider,
	logger *slog.Logger,
) *NotificationUsecase {
	return &NotificationUsecase{
		emailProvider: emailProvider,
		logger:        logger,
	}
}

// ProcessEventWithTx runs the notification logic within the consumer's DB transaction
func (u *NotificationUsecase) ProcessEventWithTx(tx *gorm.DB, messageID string, eventType string, payload map[string]interface{}) error {
	// Extract recipient based on event type
	// This is a simplified logic. In a real app, you'd have strategies per event type.
	recipient, ok := payload["user_email"].(string)
	if !ok {
		u.logger.Warn("Payload missing user_email, skipping send", slog.String("message_id", messageID))
		return errors.New("invalid payload: missing user_email")
	}

	// Process Provider
	err := u.emailProvider.Send(recipient, eventType, payload)
	
	status := domain.StatusSent
	var errMsg *string
	if err != nil {
		status = domain.StatusFailed
		e := err.Error()
		errMsg = &e
		u.logger.Error("Failed to send notification via provider", slog.String("error", e))
	}

	// 1. Save to Notification History
	notif := &domain.Notification{
		ID:               "gen-uuid", // Use UUID generator
		UserID:           "user-uuid",
		NotificationType: domain.TypeEmail,
		Provider:         domain.ProviderSendGrid,
		Recipient:        recipient,
		Status:           status,
		ErrorMessage:     errMsg,
		CreatedAt:        time.Now(),
	}

	// For simplicity, we create history directly via tx instead of repo interface
	// In strict Clean Arch, you'd pass a repo that wraps tx.
	if err := tx.Table("notification_history").Create(notif).Error; err != nil {
		return err
	}

	// 2. Publish to Outbox (notification.sent or notification.failed)
	outboxPayload, _ := json.Marshal(notif)
	outboxEvent := &domain.OutboxEventModel{
		EventType: "notification." + string(status),
		Payload:   outboxPayload,
		Status:    "pending",
	}

	if err := tx.Create(outboxEvent).Error; err != nil {
		return err
	}

	return nil
}
