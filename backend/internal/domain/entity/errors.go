package entity

import "errors"

// Domain-level sentinel errors.
// These are referenced by both services and driving adapters (HTTP handlers),
// so they live in the domain layer to avoid coupling the adapters to concrete
// service implementations.

// ErrEmptyRawText is returned when a document has no extractable text.
var ErrEmptyRawText = errors.New("categorization: raw text must not be empty")

// ErrJobAlreadyRunning is returned when trying to start a new job while one is in progress.
var ErrJobAlreadyRunning = errors.New("a batch categorization job is already running")

// ErrNothingToCategorize is returned when there are no uncategorized transactions.
var ErrNothingToCategorize = errors.New("no uncategorized transactions found")

// ErrPayslipDuplicate is returned when a payslip with the same content hash already exists.
var ErrPayslipDuplicate = errors.New("payslip already exists (duplicate)")

// ErrInvoiceDuplicate is returned when an invoice with the same content hash already exists.
var ErrInvoiceDuplicate = errors.New("invoice already exists (duplicate)")

// ErrInvoiceNotFound is returned when a referenced invoice cannot be located.
var ErrInvoiceNotFound = errors.New("invoice not found")

// ErrTransactionNotFound is returned when a referenced transaction cannot be located.
var ErrTransactionNotFound = errors.New("transaction not found")

// ErrSameAccount is returned when trying to reconcile two transactions from the same account type.
var ErrSameAccount = errors.New("source and target must be from different accounts")

// ErrInvalidCredentials is returned on failed authentication.
var ErrInvalidCredentials = errors.New("invalid username or password")

