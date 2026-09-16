CREATE TABLE units (
    id BIGSERIAL PRIMARY KEY,
    unit_number VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE invoices (
    id BIGSERIAL PRIMARY KEY,
    invoice_number VARCHAR(32) NOT NULL UNIQUE,
    unit_id BIGINT NOT NULL REFERENCES units(id),
    due_date DATE NOT NULL,
    total_amount NUMERIC(12,2) NOT NULL CHECK (total_amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE invoice_items (
    id BIGSERIAL PRIMARY KEY,
    invoice_id BIGINT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description VARCHAR(255) NOT NULL,
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0)
);

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    unit_id BIGINT NOT NULL REFERENCES units(id),
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payment_allocations (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    invoice_id BIGINT NOT NULL REFERENCES invoices(id),
    amount_allocated NUMERIC(12,2) NOT NULL CHECK (amount_allocated > 0),
    UNIQUE (payment_id, invoice_id)
);

CREATE INDEX invoices_unit_due_number_idx ON invoices (unit_id, due_date, invoice_number);
CREATE INDEX payment_allocations_invoice_id_idx ON payment_allocations (invoice_id);
