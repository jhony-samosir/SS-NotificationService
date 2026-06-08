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
	ID               string             `gorm:"primaryKey;type:uuid" json:"id"`
	UserID           string             `gorm:"index;type:uuid" json:"userId"`
	NotificationType NotificationType   `json:"notificationType"`
	Provider         ProviderType       `json:"provider"`
	Recipient        string             `json:"recipient"`
	Title            string             `json:"title"`
	Body             string             `json:"body"`
	IsRead           bool               `gorm:"default:false" json:"isRead"`
	Status           NotificationStatus `json:"status"`
	ErrorMessage     *string            `json:"errorMessage,omitempty"`
	CreatedAt        time.Time          `json:"createdAt"`
}

func (Notification) TableName() string {
	return "notification_history"
}

type NotificationRepository interface {
	Save(notification *Notification) error
}

type NotificationProvider interface {
	Send(recipient string, templateName string, data map[string]interface{}) error
}
