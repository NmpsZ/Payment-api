package repositories

import (
	"context"
	"database/sql"

	"payment-backend/modules/entities"

	"github.com/jmoiron/sqlx"
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *PaymentRepository) FindInvoicesByUnit(ctx context.Context, tx *sqlx.Tx, unitNumber string) (int64, []entities.Invoice, error) {
	var unitID int64
	err := tx.GetContext(ctx, &unitID, `SELECT id FROM units WHERE unit_number = $1`, unitNumber)
	if err == sql.ErrNoRows {
		return 0, nil, nil
	}
	if err != nil {
		return 0, nil, err
	}

	invoices := []entities.Invoice{}
	err = tx.SelectContext(ctx, &invoices,
		`SELECT i.id, i.invoice_number, i.due_date, i.total_amount,
		        COALESCE(SUM(pa.amount_allocated), 0) AS paid_amount
		 FROM invoices i
		 LEFT JOIN payment_allocations pa ON pa.invoice_id = i.id
		 WHERE i.unit_id = $1
		 GROUP BY i.id, i.invoice_number, i.due_date, i.total_amount
		 ORDER BY i.due_date ASC, i.invoice_number ASC`, unitID)
	return unitID, invoices, err
}

func (r *PaymentRepository) SavePayment(ctx context.Context, tx *sqlx.Tx, unitID int64, amount float64, allocations []entities.Allocation) (int64, error) {
	var paymentID int64
	err := tx.GetContext(ctx, &paymentID,
		`INSERT INTO payments (unit_id, amount) VALUES ($1, $2) RETURNING id`,
		unitID, amount)
	if err != nil {
		return 0, err
	}

	for _, a := range allocations {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO payment_allocations (payment_id, invoice_id, amount_allocated)
			 VALUES ($1, $2, $3)`, paymentID, a.InvoiceID, a.Amount)
		if err != nil {
			return 0, err
		}
	}

	return paymentID, nil
}
