package usecases

import (
	"context"
	"strings"

	"payment-backend/modules/entities"
	"payment-backend/modules/payments/repositories"
	"payment-backend/pkg/utils"
)

type PaymentUsecase struct {
	repo *repositories.PaymentRepository
}

func NewPaymentUsecase(repo *repositories.PaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{repo: repo}
}

type PostPaymentRequest struct {
	UnitNumber string
	Amount     float64
}

func (u *PaymentUsecase) PostPayment(ctx context.Context, req PostPaymentRequest) (*entities.PaymentResult, error) {
	if strings.TrimSpace(req.UnitNumber) == "" || req.Amount <= 0 {
		return nil, utils.ErrBadRequest("unit_number and positive amount are required")
	}

	tx, err := u.repo.BeginTx(ctx)
	if err != nil {
		return nil, utils.ErrInternal("database error")
	}
	defer tx.Rollback()

	unitID, invoices, err := u.repo.FindInvoicesByUnit(ctx, tx, req.UnitNumber)
	if err != nil {
		return nil, utils.ErrInternal("database error")
	}
	if unitID == 0 {
		return nil, utils.ErrNotFound("UNIT_NOT_FOUND", "unit was not found")
	}

	allocations, allocErr := Allocate(invoices, req.Amount)
	if allocErr == ErrNoOutstandingInvoices {
		return nil, utils.ErrUnprocessable("NO_OUTSTANDING_INVOICES", "unit has no outstanding invoices")
	}
	if allocErr == ErrOverpayment {
		return nil, utils.ErrUnprocessable("OVERPAYMENT", "payment exceeds total outstanding balance")
	}
	if allocErr != nil {
		return nil, utils.ErrInternal("could not allocate payment")
	}

	paymentID, err := u.repo.SavePayment(ctx, tx, unitID, req.Amount, allocations)
	if err != nil {
		return nil, utils.ErrInternal("could not post payment")
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.ErrInternal("could not post payment")
	}

	return &entities.PaymentResult{
		PaymentID:   paymentID,
		UnitNumber:  req.UnitNumber,
		Amount:      req.Amount,
		Allocations: allocations,
	}, nil
}
