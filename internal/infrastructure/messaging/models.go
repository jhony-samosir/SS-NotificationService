package messaging

import (
	"time"
)

// InboxEventModel ensures idempotency for incoming messages.
type InboxEventModel struct {
	MessageID   string    `gorm:"primaryKey;type:varchar(255)"`
	EventType   string    `gorm:"type:varchar(255);not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	ProcessedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	Status      string    `gorm:"type:varchar(50);not null;default:'processed'"`
}

func (InboxEventModel) TableName() string {
	return "inbox_events"
}

// OutboxEventModel stores events to be published to RabbitMQ.
type OutboxEventModel struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EventType   string    `gorm:"type:varchar(255);not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	Status      string    `gorm:"type:varchar(50);not null;default:'pending'"` // pending, published, failed
	RetryCount  int       `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (OutboxEventModel) TableName() string {
	return "outbox_events"
}
