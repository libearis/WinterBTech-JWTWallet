package wallet

import "time"

type BalanceResponse struct {
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TransferRequest struct {
	Recipient string `json:"recipient"`
	Amount    int64  `json:"amount"`
	Notes     string `json:"notes"`
}

type TransferResponse struct {
	ID           string    `json:"id"`
	Recipient    string    `json:"recipient"`
	Amount       int64     `json:"amount"`
	Notes        string    `json:"notes"`
	BalanceAfter int64     `json:"balanceAfter"`
	CreatedAt    time.Time `json:"createdAt"`
}

type TransactionItem struct {
	ID             string    `json:"id"`
	SenderEmail    string    `json:"senderEmail"`
	RecipientEmail string    `json:"recipientEmail"`
	Amount         int64     `json:"amount"`
	Type           string    `json:"type"` // "incoming" or "outgoing"
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"createdAt"`
}

type TransactionsResponse struct {
	Items      []TransactionItem `json:"items"`
	NextCursor string            `json:"nextCursor,omitempty"`
}
