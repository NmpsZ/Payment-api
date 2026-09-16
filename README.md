# Simple Billing & Payment API

A simple REST API built with **Go (Gin)** and **PostgreSQL** to create invoices and automatically pay off unpaid bills using FIFO (oldest due date first).

---

## Table of Contents

- [Background](#background)
- [Install](#install)
- [Usage](#usage)
  - [Endpoints](#endpoints)
  - [1. Create Invoice](#1-create-invoice)
  - [2. Get Invoice](#2-get-invoice)
  - [3. Make a Payment](#3-make-a-payment)
  - [Error Handling](#error-handling)
- [Architecture & Design Choices](#architecture--design-choices)
  - [Folder Structure](#folder-structure)
  - [Why PostgreSQL?](#why-postgresql)
  - [Key Decisions](#key-decisions)
- [Bonus Features](#bonus-features)
- [Testing](#testing)
- [Scaling to 10M Invoices](#scaling-to-10m-invoices)

---

## Background

Property managers (like condo or apartment offices) send monthly bills to residents for things like common fees, water, or parking. Residents often pay in parts, or they make one big payment to cover multiple months.

This API handles 3 main tasks:
1. **Create Invoice:** Takes a unit number and a list of items (e.g. water fee, repair fee). The server calculates the total price automatically.
2. **Get Invoice:** Shows bill details, total price, paid amount, remaining balance, and current status (`UNPAID`, `PARTIAL`, `PAID`).
3. **Auto-Pay (FIFO):** When a payment comes in, it pays off the oldest unpaid bill first (`due_date ASC`). If two bills have the same due date, it pays the smaller invoice number first. It supports partial payments, pays multiple bills in one go, and rejects overpayments cleanly.

---

## Install

### What you need
* **Go** (version 1.24 or higher)
* **PostgreSQL** (version 14 or higher)

### Setup Steps

1. **Copy the config file:**
   ```powershell
   cp .env.example .env
   ```
   Open `.env` and set your PostgreSQL password in `DB_PASSWORD`.

2. **Create the database and tables:**
   ```powershell
   # 1. Create the database (run once)
   psql -U postgres -c "CREATE DATABASE billing;"

   # 2. Run the SQL script to create tables
   psql -U postgres -d billing -f pkg/databases/migrations/001_initial.sql
   ```

3. **Start the server:**
   ```powershell
   go run ./app
   ```
   The API will run at `http://localhost:8080`. You can visit `GET http://localhost:8080/health` to make sure it's working.

---

## Usage

### Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/invoices` | Create a new invoice with items |
| `GET` | `/invoices/:id` | Get invoice details, current balance, and status |
| `POST` | `/payments` | Pay money to clear open invoices (oldest first) |

---

### 1. Create Invoice

Send the unit number, due date, and items. The server calculates `total_amount` for you.

* **Request:** `POST /invoices`
  ```json
  {
    "unit_number": "A101",
    "due_date": "2026-08-01",
    "items": [
      { "description": "Common fee", "amount": 1500 },
      { "description": "Water fee", "amount": 300 }
    ]
  }
  ```

* **Response:** `201 Created`
  ```json
  {
    "id": 1,
    "invoice_number": "INV-00000001",
    "unit_number": "A101",
    "due_date": "2026-08-01",
    "total_amount": 1800,
    "items": [
      { "description": "Common fee", "amount": 1500 },
      { "description": "Water fee", "amount": 300 }
    ]
  }
  ```

---

### 2. Get Invoice

Check bill details, how much has been paid, how much is left, and the status (`UNPAID`, `PARTIAL`, `PAID`).

* **Request:** `GET /invoices/1`

* **Response:** `200 OK`
  ```json
  {
    "id": 1,
    "invoice_number": "INV-00000001",
    "unit_number": "A101",
    "due_date": "2026-08-01",
    "items": [
      { "description": "Common fee", "amount": 1500 },
      { "description": "Water fee", "amount": 300 }
    ],
    "total_amount": 1800,
    "paid_amount": 1200,
    "outstanding_amount": 600,
    "status": "PARTIAL"
  }
  ```

---

### 3. Make a Payment

Pay money towards a unit. The system finds open invoices and pays them off starting from the oldest due date.

* **Request:** `POST /payments`
  ```json
  {
    "unit_number": "A101",
    "amount": 1200
  }
  ```

* **Response:** `201 Created`
  ```json
  {
    "payment_id": 1,
    "unit_number": "A101",
    "amount": 1200,
    "allocations": [
      {
        "invoice_id": 1,
        "invoice_number": "INV-00000001",
        "amount": 1200
      }
    ]
  }
  ```

---

### Error Handling

All errors return a clean and consistent JSON format:
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Friendly explanation"
  }
}
```

* **`400 Bad Request`:** Invalid input (such as missing unit number, bad date format, or negative amounts).
* **`404 Not Found`:** Unit or invoice does not exist (`UNIT_NOT_FOUND`, `INVOICE_NOT_FOUND`).
* **`422 Unprocessable Entity`:** Business rule issues:
  - `OVERPAYMENT`: The payment amount is higher than the unit's total remaining debt.
  - `NO_OUTSTANDING_INVOICES`: All bills for this unit are already fully paid.

---

## Architecture & Design Choices

### Folder Structure

The project uses a clean 3-layer pattern (Controller → Usecase → Repository):

```
Payment-backend/
├── app/
│   └── main.go                 # Starts the app (loads env, connects DB, starts server)
├── configs/
│   └── configs.go              # Loads settings from .env
├── modules/
│   ├── entities/               # Data models (Invoice, Payment)
│   ├── invoices/               # Invoices (controller, usecase, repository)
│   ├── payments/               # Payments (controller, usecase, repository)
│   └── servers/                # Gin router and error handler
├── pkg/
│   ├── databases/              # PostgreSQL connection and SQL migrations
│   └── utils/                  # Helper tools for errors and DB connection URLs
├── tests/
│   └── payments/               # Unit tests for payment allocation logic
├── .env.example                # Example configuration file
└── README.md
```

### Why PostgreSQL?

I chose **PostgreSQL** because it is a reliable relational database with great SQL and transaction support. It has complete features for managing related data and is open-source, making it suitable for both building this project and deploying to production.

### Key Decisions
1. **Rejecting Overpayment:** If someone pays more than their total debt, the API rejects it with `422 OVERPAYMENT`. Since we don't have a user wallet or credit system yet, rejecting it is the safest way to avoid lost or unaccounted money.
2. **Dynamic Status:** We don't save `status` directly in the database. Instead, it is calculated live (`total_amount - paid_amount`). This ensures the status is always 100% accurate and never gets out of sync.
3. **Auto-Create Units:** When making an invoice, if the unit doesn't exist yet, the system creates it automatically. No need to register units beforehand.
4. **Simple Numbers:** Uses standard `float64` in Go and `NUMERIC(12,2)` in PostgreSQL for clean, readable code without needing extra heavy libraries.

---

## Bonus Features

From the optional bonus list, I chose to focus on 2 items:
1. **Input Validation with Appropriate HTTP Status Codes:** Enforcing validation at the handler boundary with clear error responses (`400 Bad Request`, `404 Not Found`, and `422 Unprocessable Entity` for business rule violations).
2. **Unit Tests:** Pure in-memory unit tests covering core business logic (FIFO payment allocation rules and invoice status/balance derivation).

---

## Testing

The core payment logic (`Allocate`) is written as a pure function in memory. You can run all unit tests directly without needing a database:

```powershell
go test ./... -v
```

### 9 Test Cases Covered (9/9 PASS):
1. Pay oldest due date first (FIFO)
2. Tie-break using invoice number when due dates match
3. Partial payment (paying less than the first bill)
4. Exact payment covering multiple invoices at once
5. Skip bills that are already paid
6. Finish paying the remaining balance of a partial bill
7. Reject payments that are higher than total debt (`ErrOverpayment`)
8. Reject payments when there are no bills for the unit
9. Reject payments when all bills are already paid (`ErrNoOutstandingInvoices`)

---

## Scaling to 10M Invoices

When data grows from 10,000 to 10,000,000 invoices, I would focus on improving both the Database and Application layers:

* **Indexing:** Add indexes on frequently searched fields (such as `unit_id` and `due_date`) to keep query lookups fast.
* **Pagination:** Use pagination on listing endpoints to avoid fetching large volumes of data all at once.
* **Reduce Unnecessary Queries:** Store remaining balances directly on invoice records instead of recalculating payment totals on every read.
* **Asynchronous Processing:** Offload non-urgent tasks (like sending payment receipts or notifications) to background workers so the system uses resources efficiently and handles higher traffic smoothly.
