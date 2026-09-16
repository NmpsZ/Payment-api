package entities

import "time"

type Invoice struct {
	ID            int64     `json:"id" db:"id"`
	InvoiceNumber string    `json:"invoice_number" db:"invoice_number"`
	UnitNumber    string    `json:"unit_number" db:"unit_number"`
	DueDate       time.Time `json:"due_date" db:"due_date"`
	TotalAmount   float64   `json:"total_amount" db:"total_amount"`
	PaidAmount    float64   `json:"paid_amount" db:"paid_amount"`
}

type InvoiceItem struct {
	Description string  `json:"description" db:"description"`
	Amount      float64 `json:"amount" db:"amount"`
}

func (i Invoice) OutstandingAmount() float64 { return i.TotalAmount - i.PaidAmount }

func (i Invoice) Status() string {
	if i.PaidAmount == 0 {
		return "UNPAID"
	}
	if i.PaidAmount >= i.TotalAmount {
		return "PAID"
	}
	return "PARTIAL"
}
