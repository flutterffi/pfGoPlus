package user

import "time"

const (
	RoleAdmin  = "admin"
	RoleMember = "member"

	StatusActive   = "active"
	StatusDisabled = "disabled"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;size:64;not null"`
	DisplayName  string    `json:"display_name" gorm:"size:128;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;size:255;not null"`
	Role         string    `json:"role" gorm:"size:32;not null"`
	Status       string    `json:"status" gorm:"size:32;not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
