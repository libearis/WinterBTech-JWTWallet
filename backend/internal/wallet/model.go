package wallet

import "time"

type Wallet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Transaction struct {
	ID             string    `json:"id"`
	SenderID       string    `json:"senderId"`
	RecipientID    string    `json:"recipientId"`
	SenderEmail    string    `json:"senderEmail"`
	RecipientEmail string    `json:"recipientEmail"`
	Amount         int64     `json:"amount"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"createdAt"`
}

type IdempotencyRecord struct {
	UserID         string    `json:"userId"`
	IdempotencyKey string    `json:"idempotencyKey"`
	RequestHash    string    `json:"requestHash"`
	ResponseCode   int       `json:"responseCode"`
	ResponseBody   []byte    `json:"responseBody"`
	CreatedAt      time.Time `json:"createdAt"`
}
