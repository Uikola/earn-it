package models

import "time"

type Habit struct {
	ID               int64
	UserID           int64
	Name             string
	WeaklyGoal       int32
	RewardPerExecute int32
	CreatedAt        time.Time
}
