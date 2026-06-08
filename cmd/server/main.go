package main

import (
	"context"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/jhony-samosir/SS-NotificationService/internal/delivery/http"
	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
	"github.com/jhony-samosir/SS-NotificationService/internal/infrastructure/messaging"
	"github.com/jhony-samosir/SS-NotificationService/internal/infrastructure/provider"
	"github.com/jhony-samosir/SS-NotificationService/internal/usecase"
	"github.com/jhony-samosir/SS-NotificationService/pkg/logger"
	"github.com/jhony-samosir/SS-NotificationService/pkg/telemetry"
)

func main() {
	// Initialize Structured Logger with Trace Correlation
	log := logger.NewLogger()
	log.Info("Starting SS-NotificationService")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize OpenTelemetry
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "localhost:4317"
	}
	tp, err := telemetry.InitTracer(ctx, "ss-notification-service", otelEndpoint)
	if err != nil {
		log.Error("Failed to initialize tracer", "error", err)
	} else {
		defer tp.Shutdown(context.Background())
	}

	// 1. Initialize Database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=ss_notification_db port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		// For now we don't panic so it compiles and runs dummy
	}

	if db != nil {
		if err := db.Use(tracing.NewPlugin()); err != nil {
			log.Error("Failed to register gorm tracing plugin", "error", err)
		}
		db.AutoMigrate(&domain.InboxEventModel{}, &domain.OutboxEventModel{}, &domain.UserDeviceModel{}, &domain.Notification{})
	}

	// 2. Initialize Infrastructure Providers
	emailProv := provider.NewSendGridProvider(log, "dummy-api-key")
	smsProv := provider.NewTwilioProvider(log, "dummy-sms-key")
	pushProv := provider.NewFCMProvider(log, "dummy-fcm-key", db)

	// 3. Initialize Usecase
	notifUsecase := usecase.NewNotificationUsecase(emailProv, smsProv, pushProv, log)

	// 4. Initialize Messaging (RabbitMQ)
	rabbitMQUrl := os.Getenv("RABBITMQ_URL")
	if rabbitMQUrl == "" {
		rabbitMQUrl = "amqp://guest:guest@localhost:5672/"
	}
	
	inboxConsumer := messaging.NewInboxConsumer(rabbitMQUrl, db, notifUsecase)
	outboxWorker := messaging.NewOutboxWorker(rabbitMQUrl, db)

	// 5. Start HTTP Server for API Gateway integration
	apiUsecase := usecase.NewNotificationAPIUsecase(db)
	apiHandler := http.NewNotificationHandler(apiUsecase)

	hmacSecret := os.Getenv("GATEWAY_HMAC_SECRET")
	if hmacSecret == "" {
		hmacSecret = "default-secret"
	}
	router := http.NewRouter(hmacSecret, apiHandler)
	httpServer := &stdhttp.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Info("Starting HTTP Server on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			log.Error("HTTP server error", "error", err)
		}
	}()

	// 6. Start Messaging Workers
	go inboxConsumer.Start(ctx)
	go outboxWorker.Start(ctx)

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down servers...")
	httpServer.Shutdown(context.Background())
	log.Info("Shutdown complete")
}
