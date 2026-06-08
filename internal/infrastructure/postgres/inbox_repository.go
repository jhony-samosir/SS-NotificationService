package postgres

import (
	"database/sql"
	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
)

type InboxRepositoryImpl struct {
	db *sql.DB
}

func NewInboxRepository(db *sql.DB) domain.InboxRepository {
	return &InboxRepositoryImpl{db: db}
}

func (r *InboxRepositoryImpl) Save(event *domain.InboxEvent) error {
	// Dummy implementation
	// _, err := r.db.Exec("INSERT INTO inbox_events (message_id, type, processed_at) VALUES ($1, $2, $3)", event.MessageID, event.Type, event.ProcessedAt)
	return nil
}

func (r *InboxRepositoryImpl) Exists(messageID string) (bool, error) {
	// Dummy implementation
	// var exists bool
	// err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM inbox_events WHERE message_id = $1)", messageID).Scan(&exists)
	return false, nil
}
