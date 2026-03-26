export interface SystemInfo {
    storage_mode: string;
    db_host: string;
    db_state: string;
    version: string;
}

export interface User {
    id: string;
    username: string;
    email: string;
    full_name: string;
    address: string;
    role: string;
}

export interface Category {
    id: string;
    name: string;
    color: string;
    created_at: string;
}

export interface Vendor {
    id: string;
    name: string;
}

export interface Invoice {
    id: string;
    vendor: Vendor;
    category_id: string | null;
    amount: number;
    currency: string;
    issued_at: string;
    description: string;
    raw_text: string;
}

export interface Transaction {
    booking_date: string;
    valuta_date: string;
    description: string;
    amount: number;
    currency: string;
    type: 'credit' | 'debit';
    reference: string;
    category_id: string | null;
    content_hash: string;
    is_reconciled: boolean;
    reconciliation_id?: string;
    statement_type: 'giro' | 'credit_card' | 'extra_account';
}

export interface BankStatement {
    id: string;
    account_holder: string;
    iban: string;
    bic: string;
    account_number: string;
    statement_date: string;
    statement_no: number;
    old_balance: number;
    new_balance: number;
    currency: string;
    transactions: Transaction[];
    source_file: string;
    imported_at: string;
    content_hash: string;
    statement_type: 'giro' | 'credit_card' | 'extra_account';
}

export interface BankStatementSummary {
    id: string;
    statement_no: number;
    period_label: string;
    iban: string;
    currency: string;
    new_balance: number;
    start_date: string;
    end_date: string;
    transaction_count: number;
    statement_type: 'giro' | 'credit_card' | 'extra_account';
    has_original_file: boolean;
}

// ── Reconciliation Types ──────────────────────────────────────────────────────

export interface Reconciliation {
    id: string;
    settlement_transaction_hash: string;
    target_transaction_hash: string;
    settlement_transaction_description?: string;
    target_transaction_description?: string;
    settlement_booking_date?: string;
    target_booking_date?: string;
    settlement_statement_type?: 'giro' | 'credit_card' | 'extra_account';
    target_statement_type?: 'giro' | 'credit_card' | 'extra_account';
    amount: number;
    reconciled_at: string;
}

export interface ReconciliationPairSuggestion {
    source_transaction: Transaction;
    target_transaction: Transaction;
    match_score: number;
}

// ── Analytics Types ──────────────────────────────────────────────────────────

export interface CategoryTotal {
    category: string;
    amount: number;
    type: 'income' | 'expense';
    color: string;
}

export interface TimeSeriesPoint {
    date: string;
    income: number;
    expense: number;
}

export interface MerchantTotal {
    merchant: string;
    amount: number;
}

export interface TransactionAnalytics {
    total_income: number;
    total_expense: number;
    net_savings: number;
    category_totals: CategoryTotal[];
    time_series: TimeSeriesPoint[];
    top_merchants: MerchantTotal[];
}

// ── Batch Import Types ───────────────────────────────────────────────────────

export interface ImportResult {
    filename: string;
    status: 'imported' | 'duplicate' | 'error';
    error?: string;
    id?: string;
}

export interface ImportBatchResponse {
    summary: {
        total: number;
        imported: number;
    };
    results: ImportResult[];
}

// ── Background Job Types ─────────────────────────────────────────────────────

export interface CategorizedTransaction {
    hash: string;
    category: string;
}

export interface JobState {
    is_running: boolean;
    processed: number;
    total: number;
    status: string;
    results: CategorizedTransaction[];
}

// ── Payslip Types ────────────────────────────────────────────────────────────

export interface Bonus {
    description: string;
    amount: number;
}

export interface Payslip {
    id: string;
    original_file_name: string;
    original_file_mime?: string;   // empty / absent = JSON-imported, no binary file stored
    period_month_num: number;
    period_year: number;
    employee_name: string;
    tax_class: string;
    tax_id: string;
    gross_pay: number;
    net_pay: number;
    payout_amount: number;
    custom_deductions: number;
    bonuses: Bonus[];
    created_at: string;
}