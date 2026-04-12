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

// ErrUserNotFound is returned when a referenced user cannot be located.
var ErrUserNotFound = errors.New("user not found")

// ErrCategoryNotFound is returned when a referenced category cannot be located.
var ErrCategoryNotFound = errors.New("category not found")

// ErrBankConnectionNotFound is returned when a referenced bank connection cannot be located.
var ErrBankConnectionNotFound = errors.New("bank connection not found")

// ErrBankAccountNotFound is returned when a referenced bank account cannot be located.
var ErrBankAccountNotFound = errors.New("bank account not found")

// ErrBankStatementNotFound is returned when a referenced bank statement cannot be located.
var ErrBankStatementNotFound = errors.New("bank statement not found")

// ErrPayslipNotFound is returned when a referenced payslip cannot be located.
var ErrPayslipNotFound = errors.New("payslip not found")

// ErrSettingsNotFound is returned when a referenced setting cannot be located.
var ErrSettingsNotFound = errors.New("settings not found")

// ErrReconciliationNotFound is returned when a referenced reconciliation cannot be located.
var ErrReconciliationNotFound = errors.New("reconciliation not found")

// ErrTransactionNotFound is returned when a referenced transaction cannot be located.
var ErrTransactionNotFound = errors.New("transaction not found")

// ErrPlannedTransactionNotFound is returned when a referenced planned transaction cannot be located.
var ErrPlannedTransactionNotFound = errors.New("planned transaction not found")

// ErrInvalidPlannedTransaction is returned when a planned transaction validation fails.
var ErrInvalidPlannedTransaction = errors.New("invalid planned transaction")

// ErrSameAccount is returned when trying to reconcile two transactions from the same account type.
var ErrSameAccount = errors.New("source and target must be from different accounts")

// ErrInvalidCredentials is returned on failed authentication.
var ErrInvalidCredentials = errors.New("invalid username or password")

// ErrResetTokenInvalid is returned when a password reset token is missing, expired, or incorrect.
var ErrResetTokenInvalid = errors.New("invalid or expired reset token")

