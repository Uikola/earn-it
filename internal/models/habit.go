package models

import "time"

type Habit struct {
	ID               int64
	UserID           int64
	Name             string
	WeeklyGoal       int32
	RewardPerExecute int32
	CreatedAt        time.Time
}

type HabitLog struct {
	ID         int64
	HabitID    int64
	ExecutedAt time.Time
}
