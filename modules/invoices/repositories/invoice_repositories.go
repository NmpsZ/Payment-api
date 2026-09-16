package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"payment-backend/modules/entities"

	"github.com/jmoiron/sqlx"
)

type InvoiceRepository struct {
	db *sqlx.DB
}

func NewInvoiceRepository(db *sqlx.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

type InvoiceItemInput struct {
	Description string
	Amount      float64
}

func (r *InvoiceRepository) Create(ctx context.Context, unitNumber string, dueDate string, totalAmount float64, items []InvoiceItemInput) (int64, string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback()

	var unitID int64
	err = tx.GetContext(ctx, &unitID,
		`INSERT INTO units (unit_number) VALUES ($1)
		 ON CONFLICT (unit_number) DO UPDATE SET unit_number = EXCLUDED.unit_number
		 RETURNING id`, unitNumber)
	if err != nil {
		return 0, "", err
	}

	var invoiceID int64
	err = tx.GetContext(ctx, &invoiceID, `SELECT nextval('invoices_id_seq')`)
	if err != nil {
		return 0, "", err
	}
	invoiceNumber := fmt.Sprintf("INV-%08d", invoiceID)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO invoices (id, invoice_number, unit_id, due_date, total_amount)
		 VALUES ($1, $2, $3, $4, $5)`,
		invoiceID, invoiceNumber, unitID, dueDate, totalAmount)
	if err != nil {
		return 0, "", err
	}

	for _, item := range items {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO invoice_items (invoice_id, description, amount) VALUES ($1, $2, $3)`,
			invoiceID, item.Description, item.Amount)
		if err != nil {
			return 0, "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, "", err
	}
	return invoiceID, invoiceNumber, nil
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id int64) (*entities.Invoice, []entities.InvoiceItem, error) {
	var inv entities.Invoice
	err := r.db.GetContext(ctx, &inv,
		`SELECT i.id, i.invoice_number, u.unit_number, i.due_date, i.total_amount,
		        COALESCE(SUM(pa.amount_allocated), 0) AS paid_amount
		 FROM invoices i
		 JOIN units u ON u.id = i.unit_id
		 LEFT JOIN payment_allocations pa ON pa.invoice_id = i.id
		 WHERE i.id = $1
		 GROUP BY i.id, i.invoice_number, u.unit_number, i.due_date, i.total_amount`, id)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	items := []entities.InvoiceItem{}
	err = r.db.SelectContext(ctx, &items,
		`SELECT description, amount FROM invoice_items WHERE invoice_id = $1 ORDER BY id`, id)
	if err != nil {
		return nil, nil, err
	}

	return &inv, items, nil
}
