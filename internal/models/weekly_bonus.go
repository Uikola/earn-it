package models

import "time"

type WeeklyBonus struct {
	ID        int64
	UserID    int64
	WeekStart time.Time
	ClaimedAt time.Time
}
