package entities

type PaymentResult struct {
	PaymentID   int64        `json:"payment_id" db:"payment_id"`
	UnitNumber  string       `json:"unit_number" db:"unit_number"`
	Amount      float64      `json:"amount" db:"amount"`
	Allocations []Allocation `json:"allocations"`
}

type Allocation struct {
	InvoiceID     int64   `json:"invoice_id" db:"invoice_id"`
	InvoiceNumber string  `json:"invoice_number" db:"invoice_number"`
	Amount        float64 `json:"amount" db:"amount"`
}
