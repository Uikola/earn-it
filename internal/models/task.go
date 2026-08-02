package models

import "time"

type Task struct {
	ID            int64
	UserID        int64
	ProjectID     *int64
	Title         string
	ScheduledDate time.Time
	RewardValue   int32
	Status        string
	CompletedAt   *time.Time
	CreatedAt     time.Time
}
