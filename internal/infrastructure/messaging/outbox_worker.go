package messaging

import (
	"context"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type OutboxWorker struct {
	url  string
	db   *gorm.DB
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewOutboxWorker(url string, db *gorm.DB) *OutboxWorker {
	return &OutboxWorker{
		url: url,
		db:  db,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) error {
	for {
		err := w.connectAndWork(ctx)
		if err != nil {
			slog.Error("Outbox worker error", "error", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
			slog.Info("Reconnecting outbox worker...")
		}
	}
}

func (w *OutboxWorker) connectAndWork(ctx context.Context) error {
	conn, err := amqp.Dial(w.url)
	if err != nil {
		return err
	}
	defer conn.Close()
	w.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	w.ch = ch

	err = ch.ExchangeDeclare("samstore.events", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Outbox worker shutting down")
			return nil
		case <-ticker.C:
			w.pollAndPublish(ctx)
		}
	}
}

func (w *OutboxWorker) pollAndPublish(ctx context.Context) {
	var events []OutboxEventModel
	// Fetch pending events
	if err := w.db.WithContext(ctx).Where("status = ?", "pending").Limit(50).Find(&events).Error; err != nil {
		slog.Error("Failed to fetch outbox events", "error", err)
		return
	}

	for _, event := range events {
		err := w.ch.PublishWithContext(ctx,
			"samstore.events", // exchange
			event.EventType,   // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				MessageId:    event.ID,
				Body:         event.Payload,
			},
		)

		if err != nil {
			slog.Error("Failed to publish outbox event", "event_id", event.ID, "error", err)
			// Increment retry count
			w.db.WithContext(ctx).Model(&event).Update("retry_count", event.RetryCount+1)
			continue
		}

		// Mark as published
		w.db.WithContext(ctx).Model(&event).Update("status", "published")
		slog.Debug("Outbox event published successfully", "event_id", event.ID)
	}
}
