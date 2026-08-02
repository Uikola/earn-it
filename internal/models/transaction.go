package models

import "time"

type Transaction struct {
	ID        int64
	UserID    int64
	Amount    int32
	Source    string
	SourceID  int64
	CreatedAt time.Time
}
