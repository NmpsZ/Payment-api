package usecases

import (
	"errors"
	"sort"

	"payment-backend/modules/entities"
)

var ErrNoOutstandingInvoices = errors.New("unit has no outstanding invoices")
var ErrOverpayment = errors.New("payment amount exceeds total outstanding balance")

func Allocate(invoices []entities.Invoice, paymentAmount float64) ([]entities.Allocation, error) {
	if paymentAmount <= 0 {
		return nil, errors.New("payment amount must be greater than zero")
	}

	pendingInvoices := make([]entities.Invoice, 0, len(invoices))
	totalOutstanding := 0.0

	for _, inv := range invoices {
		outstanding := inv.OutstandingAmount()
		if outstanding > 0 {
			pendingInvoices = append(pendingInvoices, inv)
			totalOutstanding += outstanding
		}
	}

	if len(pendingInvoices) == 0 {
		return nil, ErrNoOutstandingInvoices
	}

	if paymentAmount > totalOutstanding {
		return nil, ErrOverpayment
	}

	// Sort oldest due date first; tie-break by invoice number ascending
	sort.Slice(pendingInvoices, func(i, j int) bool {
		if pendingInvoices[i].DueDate.Equal(pendingInvoices[j].DueDate) {
			return pendingInvoices[i].InvoiceNumber < pendingInvoices[j].InvoiceNumber
		}
		return pendingInvoices[i].DueDate.Before(pendingInvoices[j].DueDate)
	})

	remainingPayment := paymentAmount
	allocations := []entities.Allocation{}

	for _, inv := range pendingInvoices {
		if remainingPayment <= 0 {
			break
		}

		amountToCut := remainingPayment
		if outstanding := inv.OutstandingAmount(); outstanding < amountToCut {
			amountToCut = outstanding
		}

		allocations = append(allocations, entities.Allocation{
			InvoiceID:     inv.ID,
			InvoiceNumber: inv.InvoiceNumber,
			Amount:        amountToCut,
		})

		remainingPayment -= amountToCut
	}

	return allocations, nil
}
