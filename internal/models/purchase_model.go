package models

improt ()

type Purchase struct {
	UserID    string    `json:"user_id"`  
	MerchID   int       `json:"merch_id"`
	Quantity  int       `json:"quantity"`
	Purchased time.Time `json:"purchased_at"`
}