package role

import "time"

type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;size:64;not null"`
	DisplayName string    `json:"display_name" gorm:"size:128;not null"`
	Permissions string    `json:"permissions" gorm:"size:1024;not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Role) TableName() string {
	return "roles"
}
