package invoices_test

import (
	"testing"
	"time"

	"payment-backend/modules/entities"
)

func TestInvoiceEntity_Status(t *testing.T) {
	tests := []struct {
		name        string
		total       float64
		paid        float64
		expected    string
		outstanding float64
	}{
		{
			name:        "unpaid when paid is 0",
			total:       1500,
			paid:        0,
			expected:    "UNPAID",
			outstanding: 1500,
		},
		{
			name:        "partial when partially paid",
			total:       1500,
			paid:        500,
			expected:    "PARTIAL",
			outstanding: 1000,
		},
		{
			name:        "paid when fully paid",
			total:       1500,
			paid:        1500,
			expected:    "PAID",
			outstanding: 0,
		},
		{
			name:        "paid when overpaid edge case",
			total:       1500,
			paid:        2000,
			expected:    "PAID",
			outstanding: -500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := entities.Invoice{
				ID:            1,
				InvoiceNumber: "INV-00000001",
				UnitNumber:    "A101",
				DueDate:       time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
				TotalAmount:   tt.total,
				PaidAmount:    tt.paid,
			}

			if got := inv.Status(); got != tt.expected {
				t.Fatalf("expected status %s, got %s", tt.expected, got)
			}
			if got := inv.OutstandingAmount(); got != tt.outstanding {
				t.Fatalf("expected outstanding %.2f, got %.2f", tt.outstanding, got)
			}
		})
	}
}
