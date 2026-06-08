package domain

import (
	"time"
)

type NotificationType string
type ProviderType string
type NotificationStatus string

const (
	TypeEmail NotificationType = "EMAIL"
	TypeSMS   NotificationType = "SMS"
	TypePush  NotificationType = "PUSH"

	ProviderSendGrid ProviderType = "SENDGRID"
	ProviderTwilio   ProviderType = "TWILIO"
	ProviderFCM      ProviderType = "FCM"

	StatusSent    NotificationStatus = "SENT"
	StatusFailed  NotificationStatus = "FAILED"
	StatusBounced NotificationStatus = "BOUNCED"
)

type Notification struct {
	ID               string
	UserID           string
	NotificationType NotificationType
	Provider         ProviderType
	Recipient        string
	Status           NotificationStatus
	ErrorMessage     *string
	CreatedAt        time.Time
}

type NotificationRepository interface {
	Save(notification *Notification) error
}

type NotificationProvider interface {
	Send(recipient string, templateName string, data map[string]interface{}) error
}
