package domain

import "time"

type InboxEvent struct {
	MessageID   string
	Type        string
	ProcessedAt time.Time
}

type InboxRepository interface {
	Save(event *InboxEvent) error
	Exists(messageID string) (bool, error)
}
