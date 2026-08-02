package models

import "time"

type ShopItem struct {
	ID          int64
	UserID      int64
	Name        string
	Price       int32
	IsPurchased bool
	CreatedAt   time.Time
}

type Purchase struct {
	ID          int64
	UserID      int64
	ShopItemID  int64
	PricePaid   int32
	PurchasedAt time.Time
}
