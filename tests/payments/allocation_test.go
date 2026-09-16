package payments_test

import (
	"testing"
	"time"

	"payment-backend/modules/entities"
	"payment-backend/modules/payments/usecases"
)

func date(s string) time.Time { t, _ := time.Parse("2006-01-02", s); return t }

func TestAllocate(t *testing.T) {
	tests := []struct {
		name      string
		invoices  []entities.Invoice
		amount    float64
		wantErr   error
		wantAlloc []entities.Allocation
	}{
		{
			name: "oldest due date first",
			invoices: []entities.Invoice{
				{ID: 2, InvoiceNumber: "INV-002", DueDate: date("2026-09-01"), TotalAmount: 500},
				{ID: 1, InvoiceNumber: "INV-001", DueDate: date("2026-08-01"), TotalAmount: 1000},
			},
			amount: 1200,
			wantAlloc: []entities.Allocation{
				{InvoiceID: 1, InvoiceNumber: "INV-001", Amount: 1000},
				{InvoiceID: 2, InvoiceNumber: "INV-002", Amount: 200},
			},
		},
		{
			name: "tie-break by invoice number when same due date",
			invoices: []entities.Invoice{
				{ID: 3, InvoiceNumber: "INV-003", DueDate: date("2026-08-01"), TotalAmount: 300},
				{ID: 1, InvoiceNumber: "INV-001", DueDate: date("2026-08-01"), TotalAmount: 200},
				{ID: 2, InvoiceNumber: "INV-002", DueDate: date("2026-08-01"), TotalAmount: 400},
			},
			amount: 500,
			wantAlloc: []entities.Allocation{
				{InvoiceID: 1, InvoiceNumber: "INV-001", Amount: 200},
				{InvoiceID: 2, InvoiceNumber: "INV-002", Amount: 300},
			},
		},
		{
			name: "partial payment smaller than first invoice",
			invoices: []entities.Invoice{
				{ID: 1, InvoiceNumber: "INV-001", DueDate: date("2026-08-01"), TotalAmount: 1000},
			},
			amount: 250,
			wantAlloc: []entities.Allocation{
				{InvoiceID: 1, InvoiceNumber: "INV-001", Amount: 250},
			},
		},
		{
			name: "exact full payment across multiple invoices",
			invoices: []entities.Invoice{
				{ID: 1, InvoiceNumber: "INV-001", DueDate: date("2026-08-01"), TotalAmount: 500},
				{ID: 2, InvoiceNumber: "INV-002", DueDate: date("2026-09-01"), TotalAmount: 500},
			},
			amount: 1000,
			wantAlloc: []entities.Allocation{
				{InvoiceID: 1, InvoiceNumber: "INV-001", Amount: 500},
				{InvoiceID: 2, InvoiceNumber: "INV-002", Amount: 500},
			},
		},
		{
			name: "already-paid invoices are skipped",
			invoices: []entities.Invoice{
				{ID: 1, InvoiceNumber: "INV-001", DueDate: date("2026-08-01"), TotalAmount: 500, PaidAmount: 500},
				{ID: 2, InvoiceNumber: "INV-002", DueDate: date("2026-09-01"), TotalAmount: 300},
			},
			amount: 300,
			wantAlloc: []entities.Allocation{
				{InvoiceID: 2, InvoiceNumber: "INV-002", Amount: 300},
			},
		},
		{
			name: "partially-paid invoice gets remaining allocated",
			invoices: []entities.Invoice{
				{ID: 1, InvoiceNumber: "INV-001", DueDate: date("2026-08-01"), TotalAmount: 1000, PaidAmount: 600},
			},
			amount: 400,
			wantAlloc: []entities.Allocation{
				{InvoiceID: 1, InvoiceNumber: "INV-001", Amount: 400},
			},
		},
		{
			name: "overpayment rejected",
			invoices: []entities.Invoice{
				{ID: 1, InvoiceNumber: "INV-001", DueDate: date("2026-08-01"), TotalAmount: 500},
			},
			amount:  600,
			wantErr: usecases.ErrOverpayment,
		},
		{
			name:     "no outstanding invoices rejected (empty list)",
			invoices: []entities.Invoice{},
			amount:   100,
			wantErr:  usecases.ErrNoOutstandingInvoices,
		},
		{
			name: "all invoices fully paid treated as no outstanding",
			invoices: []entities.Invoice{
				{ID: 1, InvoiceNumber: "INV-001", DueDate: date("2026-08-01"), TotalAmount: 500, PaidAmount: 500},
			},
			amount:  100,
			wantErr: usecases.ErrNoOutstandingInvoices,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := usecases.Allocate(tt.invoices, tt.amount)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.wantAlloc) {
				t.Fatalf("expected %d allocations, got %d: %+v", len(tt.wantAlloc), len(got), got)
			}
			for i, want := range tt.wantAlloc {
				if got[i].InvoiceID != want.InvoiceID {
					t.Errorf("allocation[%d] invoice_id: want %d, got %d", i, want.InvoiceID, got[i].InvoiceID)
				}
				if got[i].Amount != want.Amount {
					t.Errorf("allocation[%d] amount: want %v, got %v", i, want.Amount, got[i].Amount)
				}
			}
		})
	}
}
