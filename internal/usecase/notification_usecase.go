package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
	"github.com/google/uuid"

	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
)

type NotificationUsecase struct {
	emailProvider domain.NotificationProvider
	smsProvider   domain.NotificationProvider
	pushProvider  domain.NotificationProvider
	logger        *slog.Logger
}

func NewNotificationUsecase(
	emailProvider domain.NotificationProvider,
	smsProvider domain.NotificationProvider,
	pushProvider domain.NotificationProvider,
	logger *slog.Logger,
) *NotificationUsecase {
	return &NotificationUsecase{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
		pushProvider:  pushProvider,
		logger:        logger,
	}
}

// ProcessEventWithTx runs the notification logic within the consumer's DB transaction
func (u *NotificationUsecase) ProcessEventWithTx(ctx context.Context, tx *gorm.DB, messageID string, eventType string, payload map[string]interface{}) error {
	tracer := otel.Tracer("notification-usecase")
	ctx, span := tracer.Start(ctx, "process_event")
	defer span.End()

	// Check for channels
	email, hasEmail := payload["user_email"].(string)
	phone, hasPhone := payload["user_phone"].(string)

	if !hasEmail && !hasPhone {
		u.logger.WarnContext(ctx, "Payload missing recipient details", slog.String("message_id", messageID))
		return errors.New("invalid payload: missing recipient")
	}

	status := domain.StatusSent
	var errMsg *string
	var lastErr error

	// Retry wrapper
	sendWithRetry := func(provider domain.NotificationProvider, recipient string) error {
		var err error
		maxRetries := 3
		backoff := 1 * time.Second

		for i := 0; i <= maxRetries; i++ {
			err = provider.Send(recipient, eventType, payload)
			if err == nil {
				return nil
			}
			if i < maxRetries {
				u.logger.WarnContext(ctx, "Provider send failed, retrying", slog.Int("attempt", i+1), slog.Duration("backoff", backoff))
				time.Sleep(backoff)
				backoff *= 2
			}
		}
		return err
	}

	if hasEmail {
		if err := sendWithRetry(u.emailProvider, email); err != nil {
			lastErr = err
		}
	}

	if hasPhone {
		if err := sendWithRetry(u.smsProvider, phone); err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		status = domain.StatusFailed
		e := lastErr.Error()
		errMsg = &e
		u.logger.ErrorContext(ctx, "Failed to send notification after retries", slog.String("error", e))
	}

	userID, _ := payload["user_id"].(string)
	title, _ := payload["title"].(string)
	body, _ := payload["body"].(string)

	if userID == "" {
		userID = "system"
	}
	if title == "" {
		title = "New Notification"
	}

	// 1. Save to Notification History
	notif := &domain.Notification{
		ID:               uuid.NewString(),
		UserID:           userID,
		Title:            title,
		Body:             body,
		IsRead:           false,
		NotificationType: domain.TypeEmail,
		Provider:         domain.ProviderSendGrid,
		Recipient:        "multiple",
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
