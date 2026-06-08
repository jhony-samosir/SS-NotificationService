package main

import (
	"context"
	"log/slog"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
	"github.com/jhony-samosir/SS-NotificationService/internal/infrastructure/messaging"
	"github.com/jhony-samosir/SS-NotificationService/internal/infrastructure/provider"
	"github.com/jhony-samosir/SS-NotificationService/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Starting SS-NotificationService")

	// 1. Initialize Database (Dummy connection string for bootstrap)
	dsn := "host=localhost user=postgres password=postgres dbname=ss_notification_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		// For now we don't panic so it compiles and runs dummy
	}

	if db != nil {
		db.AutoMigrate(&domain.InboxEventModel{}, &domain.OutboxEventModel{})
	}

	// 2. Initialize Infrastructure Providers
	emailProv := provider.NewSendGridProvider(logger, "dummy-api-key")

	// 3. Initialize Usecase
	notifUsecase := usecase.NewNotificationUsecase(emailProv, logger)

	// 4. Initialize Messaging (RabbitMQ)
	rabbitMQUrl := "amqp://guest:guest@localhost:5672/"
	
	inboxConsumer := messaging.NewInboxConsumer(rabbitMQUrl, db, notifUsecase)
	outboxWorker := messaging.NewOutboxWorker(rabbitMQUrl, db)

	// 5. Start Workers
	ctx := context.Background()
	
	go inboxConsumer.Start(ctx)
	go outboxWorker.Start(ctx)

	// Block forever
	select {}
}
