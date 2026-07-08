package audit

import "time"

const (
	StatusSuccess = "success"
	StatusFailure = "failure"
)

type Log struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ActorID       uint      `json:"actor_id"`
	ActorUsername string    `json:"actor_username" gorm:"size:64;not null"`
	Action        string    `json:"action" gorm:"size:64;not null"`
	Resource      string    `json:"resource" gorm:"size:64;not null"`
	ResourceID    string    `json:"resource_id" gorm:"size:64"`
	Status        string    `json:"status" gorm:"size:32;not null"`
	TraceID       string    `json:"trace_id" gorm:"size:128"`
	Detail        string    `json:"detail" gorm:"size:512"`
	CreatedAt     time.Time `json:"created_at"`
}

func (Log) TableName() string {
	return "audit_logs"
}
