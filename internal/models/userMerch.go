package models

import "time"

type UserMerch struct {
	MerchID     int       `json:"merch_id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	PurchasedAt time.Time `json:"purchased_at"`
}