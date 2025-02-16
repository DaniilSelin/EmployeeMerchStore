package models

import "time"

type Ledger struct {
	ID          int       `json:"id"`
	UserID      string    `json:"user_id"`
	MovementType string   `json:"movement_type"` // Тип движения, например, 'transfer_in', 'transfer_out', 'purchase'
	Amount      float64   `json:"amount"`
	ReferenceID *int      `json:"reference_id,omitempty"` 
	Reference_id_usr string `json:"reference_id_usr,omitempty"`// "Костыль"
	CreatedAt   time.Time `json:"created_at"`
}
