package models

import (
	"time"
)

type User struct {
	ID                int64
	Timezone          string
	Balance           int32
	RewardWeeklyBonus int32
	CreatedAt         time.Time
}
