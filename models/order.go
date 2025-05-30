package models

type Order struct {
    ID               int64   `json:"id"`
    Symbol           string  `json:"symbol"`
    Side             string  `json:"side"`   // "buy" or "sell"
    Type             string  `json:"type"`   // "limit" or "market"
    Price            float64 `json:"price,omitempty"`
    InitialQuantity  int     `json:"initial_quantity"`
    RemainingQuantity int    `json:"remaining_quantity"`
    Status           string  `json:"status"`
    CreatedAt        string  `json:"created_at"`
}
