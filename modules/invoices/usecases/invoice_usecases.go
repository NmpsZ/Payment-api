package usecases

import (
	"context"
	"strings"
	"time"

	"payment-backend/modules/entities"
	"payment-backend/modules/invoices/repositories"
	"payment-backend/pkg/utils"
)

type InvoiceUsecase struct {
	repo *repositories.InvoiceRepository
}

func NewInvoiceUsecase(repo *repositories.InvoiceRepository) *InvoiceUsecase {
	return &InvoiceUsecase{repo: repo}
}

type CreateInvoiceRequest struct {
	UnitNumber string
	DueDate    string
	Items      []CreateInvoiceItem
}

type CreateInvoiceItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type CreateInvoiceResponse struct {
	ID            int64               `json:"id"`
	InvoiceNumber string              `json:"invoice_number"`
	UnitNumber    string              `json:"unit_number"`
	DueDate       string              `json:"due_date"`
	TotalAmount   float64             `json:"total_amount"`
	Items         []CreateInvoiceItem `json:"items"`
}

type GetInvoiceResponse struct {
	ID                int64                  `json:"id"`
	InvoiceNumber     string                 `json:"invoice_number"`
	UnitNumber        string                 `json:"unit_number"`
	DueDate           string                 `json:"due_date"`
	Items             []entities.InvoiceItem `json:"items"`
	TotalAmount       float64                `json:"total_amount"`
	PaidAmount        float64                `json:"paid_amount"`
	OutstandingAmount float64                `json:"outstanding_amount"`
	Status            string                 `json:"status"`
}

func (u *InvoiceUsecase) Create(ctx context.Context, req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	if strings.TrimSpace(req.UnitNumber) == "" || len(req.Items) == 0 {
		return nil, utils.ErrBadRequest("unit_number and items are required")
	}

	_, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return nil, utils.ErrBadRequest("due_date must use YYYY-MM-DD")
	}

	total := 0.0
	repoItems := make([]repositories.InvoiceItemInput, len(req.Items))
	for i, item := range req.Items {
		if strings.TrimSpace(item.Description) == "" || item.Amount <= 0 {
			return nil, utils.ErrBadRequest("each item needs a description and positive amount")
		}
		total += item.Amount
		repoItems[i] = repositories.InvoiceItemInput{
			Description: item.Description,
			Amount:      item.Amount,
		}
	}

	invoiceID, invoiceNumber, err := u.repo.Create(ctx, req.UnitNumber, req.DueDate, total, repoItems)
	if err != nil {
		return nil, utils.ErrInternal("could not create invoice")
	}

	return &CreateInvoiceResponse{
		ID:            invoiceID,
		InvoiceNumber: invoiceNumber,
		UnitNumber:    req.UnitNumber,
		DueDate:       req.DueDate,
		TotalAmount:   total,
		Items:         req.Items,
	}, nil
}

func (u *InvoiceUsecase) GetByID(ctx context.Context, id int64) (*GetInvoiceResponse, error) {
	if id < 1 {
		return nil, utils.ErrBadRequest("invoice id must be a positive integer")
	}

	inv, items, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, utils.ErrInternal("database error")
	}
	if inv == nil {
		return nil, utils.ErrNotFound("INVOICE_NOT_FOUND", "invoice was not found")
	}

	return &GetInvoiceResponse{
		ID:                inv.ID,
		InvoiceNumber:     inv.InvoiceNumber,
		UnitNumber:        inv.UnitNumber,
		DueDate:           inv.DueDate.Format("2006-01-02"),
		Items:             items,
		TotalAmount:       inv.TotalAmount,
		PaidAmount:        inv.PaidAmount,
		OutstandingAmount: inv.OutstandingAmount(),
		Status:            inv.Status(),
	}, nil
}
