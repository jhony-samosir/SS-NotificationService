package messaging

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
	"github.com/jhony-samosir/SS-NotificationService/internal/usecase"
)

type InboxConsumer struct {
	url     string
	queue   string
	db      *gorm.DB
	usecase *usecase.NotificationUsecase
	conn    *amqp.Connection
	ch      *amqp.Channel
}

func NewInboxConsumer(url string, db *gorm.DB, uc *usecase.NotificationUsecase) *InboxConsumer {
	return &InboxConsumer{
		url:     url,
		queue:   "notification-service.events",
		db:      db,
		usecase: uc,
	}
}

func (c *InboxConsumer) Start(ctx context.Context) error {
	for {
		err := c.consume(ctx)
		if err != nil {
			slog.Error("Inbox consumer error", "error", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
			slog.Info("Reconnecting inbox consumer...")
		}
	}
}

func (c *InboxConsumer) consume(ctx context.Context) error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}
	defer conn.Close()
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	c.ch = ch

	err = ch.ExchangeDeclare("samstore.events", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(c.queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Bind to multiple events
	routingKeys := []string{"order.created", "payment.success", "auth.user.registered"}
	for _, rk := range routingKeys {
		err = ch.QueueBind(q.Name, rk, "samstore.events", false, nil)
		if err != nil {
			return err
		}
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	slog.Info("Inbox consumer started", "queue", q.Name)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Inbox consumer shutting down")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			c.processMessage(ctx, msg)
		}
	}
}

func (c *InboxConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	messageID := msg.MessageId
	if messageID == "" {
		slog.Warn("Received message without MessageId", "routing_key", msg.RoutingKey)
		msg.Reject(false)
		return
	}

	eventType := msg.RoutingKey
	slog.Info("Processing incoming event", "message_id", messageID, "routing_key", eventType)

	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Idempotency check with ON CONFLICT DO NOTHING
		inboxEvent := &domain.InboxEventModel{
			MessageID: messageID,
			EventType: eventType,
			Payload:   msg.Body,
			Status:    "processed",
		}

		err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(inboxEvent).Error
		if err != nil {
			return err
		}

		if tx.RowsAffected == 0 {
			slog.Debug("Message already processed (idempotent)", "message_id", messageID)
			return nil
		}

		// Parse payload
		var payload map[string]interface{}
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			return err
		}

		// Pass the transaction to the usecase so it can insert to notification_history
		// and outbox_events in the same transaction
		return c.usecase.ProcessEventWithTx(tx, messageID, eventType, payload)
	})

	if err != nil {
		slog.Error("Failed to process message", "message_id", messageID, "error", err)
		msg.Nack(false, true)
	} else {
		msg.Ack(false)
	}
}
