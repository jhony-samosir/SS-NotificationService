package domain

import "time"

// UserDeviceModel maps a user to their FCM device token
type UserDeviceModel struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string    `gorm:"type:varchar(255);not null;index"`
	DeviceToken string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	DeviceType  string    `gorm:"type:varchar(50);not null"` // "ios", "android", "web"
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (UserDeviceModel) TableName() string {
	return "user_devices"
}
