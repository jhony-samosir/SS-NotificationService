package usecase

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
)

type NotificationAPIUsecase struct {
	db *gorm.DB
}

func NewNotificationAPIUsecase(db *gorm.DB) *NotificationAPIUsecase {
	return &NotificationAPIUsecase{db: db}
}

func (u *NotificationAPIUsecase) RegisterDevice(ctx context.Context, userID, token, deviceType string) error {
	device := &domain.UserDeviceModel{
		UserID:      userID,
		DeviceToken: token,
		DeviceType:  deviceType,
	}

	// Upsert: If token exists, just update user_id and device_type
	return u.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "device_token"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "device_type", "updated_at"}),
	}).Create(device).Error
}

func (u *NotificationAPIUsecase) GetHistory(ctx context.Context, userID string, page, limit int) ([]domain.Notification, error) {
	var history []domain.Notification
	offset := (page - 1) * limit
	err := u.db.WithContext(ctx).Table("notification_history").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&history).Error
	return history, err
}

func (u *NotificationAPIUsecase) MarkAsRead(ctx context.Context, userID, notificationID string) error {
	return u.db.WithContext(ctx).Table("notification_history").
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

func (u *NotificationAPIUsecase) MarkAllAsRead(ctx context.Context, userID string) error {
	return u.db.WithContext(ctx).Table("notification_history").
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
