package rabbitmq

import (
	"context"
	"log/slog"

	"github.com/jhony-samosir/SS-NotificationService/internal/usecase"
)

type Consumer struct {
	usecase *usecase.NotificationUsecase
	logger  *slog.Logger
}

func NewConsumer(uc *usecase.NotificationUsecase, logger *slog.Logger) *Consumer {
	return &Consumer{
		usecase: uc,
		logger:  logger,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("Starting RabbitMQ Consumer...")
	// Dummy implementation
	// In reality, this would connect to amqp091-go, declare queues, bind to samstore.events
	// and consume messages in a goroutine, invoking c.usecase.ProcessOrderCreatedEvent.
}
