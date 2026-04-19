import type { Transaction } from './transaction';

export interface BankStatement {
    id: string;
    account_holder: string;
    iban: string;
    statement_date: string;
    statement_no: number;
    old_balance: number;
    new_balance: number;
    currency: string;
    transactions: Transaction[];
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

export type ConnectionStatus = 'initialized' | 'linked' | 'expired' | 'failed';

export interface BankAccount {
    id: string;
    connection_id: string;
    provider_account_id: string;
    iban: string;
    name: string;
    currency: string;
    balance: number;
    last_synced_at: string;
    account_type: 'giro' | 'credit_card' | 'extra_account';
}

export interface BankConnection {
    id: string;
    user_id: string;
    provider: string;
    institution_id: string;
    institution_name: string;
    requisition_id: string;
    reference_id: string;
    status: ConnectionStatus;
    auth_link?: string;
    created_at: string;
    expires_at: string | null;
    accounts?: BankAccount[];
}

export interface BankInstitution {
    id: string;
    name: string;
    bic: string;
    logo?: string;
    country: string;
}