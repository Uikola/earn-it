package models

import "time"

type Transaction struct {
	ID         int64
	UserID     int64
	Amount     int32
	Source     string
	SourceName string
	CreatedAt  time.Time
}

type TransactionWithDetails struct {
	ID         int64
	Amount     int32
	Source     string
	SourceName string
	CreatedAt  time.Time
}
