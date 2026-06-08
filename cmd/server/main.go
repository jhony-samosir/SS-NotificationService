package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/jhony-samosir/SS-NotificationService/internal/delivery/rabbitmq"
	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
	"github.com/jhony-samosir/SS-NotificationService/internal/infrastructure/postgres"
	"github.com/jhony-samosir/SS-NotificationService/internal/infrastructure/provider"
	"github.com/jhony-samosir/SS-NotificationService/internal/usecase"
)

// Dummy Notification Repo to satisfy interface
type dummyNotifRepo struct{}

func (d *dummyNotifRepo) Save(notification *domain.Notification) error {
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Starting SS-NotificationService")

	// Dummy DB
	var db *sql.DB

	// Infrastructure
	inboxRepo := postgres.NewInboxRepository(db)
	notifRepo := &dummyNotifRepo{}
	emailProv := provider.NewSendGridProvider(logger, "dummy-api-key")

	// Usecase
	notifUsecase := usecase.NewNotificationUsecase(inboxRepo, notifRepo, emailProv, logger)

	// Delivery
	consumer := rabbitmq.NewConsumer(notifUsecase, logger)
	consumer.Start(context.Background())

	// Block forever
	select {}
}
