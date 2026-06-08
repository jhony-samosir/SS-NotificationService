package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
)

type NotificationUsecase struct {
	inboxRepo    domain.InboxRepository
	notifRepo    domain.NotificationRepository
	emailProvider domain.NotificationProvider
	logger       *slog.Logger
}

func NewNotificationUsecase(
	inboxRepo domain.InboxRepository,
	notifRepo domain.NotificationRepository,
	emailProvider domain.NotificationProvider,
	logger *slog.Logger,
) *NotificationUsecase {
	return &NotificationUsecase{
		inboxRepo:    inboxRepo,
		notifRepo:    notifRepo,
		emailProvider: emailProvider,
		logger:       logger,
	}
}

func (u *NotificationUsecase) ProcessOrderCreatedEvent(ctx context.Context, messageID string, payload map[string]interface{}) error {
	// 1. Idempotency Check
	exists, err := u.inboxRepo.Exists(messageID)
	if err != nil {
		return err
	}
	if exists {
		u.logger.Info("Message already processed, skipping", slog.String("message_id", messageID))
		return nil
	}

	// 2. Process Notification
	recipient, ok := payload["user_email"].(string)
	if !ok {
		return errors.New("invalid payload: missing user_email")
	}

	err = u.emailProvider.Send(recipient, "order_created", payload)
	
	status := domain.StatusSent
	var errMsg *string
	if err != nil {
		status = domain.StatusFailed
		e := err.Error()
		errMsg = &e
		u.logger.Error("Failed to send email", slog.String("error", e))
	}

	// 3. Save History
	notif := &domain.Notification{
		ID:               "gen-uuid", // In a real app, generate UUID
		UserID:           "user-uuid", // Extract from payload
		NotificationType: domain.TypeEmail,
		Provider:         domain.ProviderSendGrid,
		Recipient:        recipient,
		Status:           status,
		ErrorMessage:     errMsg,
		CreatedAt:        time.Now(),
	}
	_ = u.notifRepo.Save(notif)

	// 4. Save Inbox Event to mark as processed
	if err == nil {
		err = u.inboxRepo.Save(&domain.InboxEvent{
			MessageID:   messageID,
			Type:        "order.created",
			ProcessedAt: time.Now(),
		})
		if err != nil {
			return err
		}
	}

	return err
}
