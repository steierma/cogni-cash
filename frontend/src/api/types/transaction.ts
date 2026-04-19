export interface Transaction {
    id: string;
    user_id: string;
    booking_date: string;
    valuta_date: string;
    description: string;
    counterparty_name?: string;
    counterparty_iban?: string;
    bank_transaction_code?: string;
    mandate_reference?: string;
    amount: number;
    currency: string;
    type: 'credit' | 'debit';
    reference: string;
    category_id: string | null;
    content_hash: string;
    is_reconciled: boolean;
    reconciliation_id?: string;
    reviewed: boolean;
    statement_type: 'giro' | 'credit_card' | 'extra_account';
    bank_statement_id?: string | null;
    location?: string;
    is_prediction?: boolean;
    skip_forecasting?: boolean;
    is_shared: boolean;
    owner_id?: string;
    subscription_id?: string | null;
}

// ── Analytics ──────────────────────────────────────────────────────────────
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
    category_amounts: Record<string, number>;
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

// ── Forecasting ────────────────────────────────────────────────────────────
export interface ForecastPoint {
    date: string;
    expected_balance: number;
    income: number;
    expense: number;
    category_amounts: Record<string, number>;
}

export interface PredictedTransaction extends Transaction {
    probability: number;
}

export interface PatternExclusion {
    id: string;
    user_id: string;
    match_term: string;
    created_at: string;
}

export interface CashFlowForecast {
    current_balance: number;
    time_series: ForecastPoint[];
    predictions: PredictedTransaction[];
}

// ── Planned Transactions ───────────────────────────────────────────────────
export interface PlannedTransaction {
    id: string;
    user_id: string;
    amount: number;
    date: string;
    description: string;
    category_id: string;
    status: 'pending' | 'matched' | 'expired';
    matched_transaction_id?: string;
    interval_months: number;
    end_date?: string;
    is_superseded: boolean;
    created_at: string;
}

export interface CreatePlannedTransactionRequest {
    amount: number;
    date: string;
    description: string;
    category_id: string | null;
    interval_months?: number;
    end_date?: string;
}

export interface UpdatePlannedTransactionRequest {
    amount?: number;
    date?: string;
    description?: string;
    category_id?: string | null;
    interval_months?: number;
    end_date?: string;
    status?: 'pending' | 'matched' | 'expired';
}

// ── Reconciliations ────────────────────────────────────────────────────────
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

// ── Background Jobs ────────────────────────────────────────────────────────
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