import type { Invoice } from './invoice';

export interface Category {
    id: string;
    name: string;
    color: string;
    is_variable_spending: boolean;
    forecast_strategy: string;
    is_shared: boolean;
    shared_with?: string[];
    owner_id: string;
    created_at: string;
    deleted_at?: string;
}

export interface UserSpending {
    user_id: string;
    username: string;
    amount: number;
}

export interface CategoryBalance {
    category_id: string;
    category_name: string;
    total_spent: number;
    user_breakdown: UserSpending[];
}

export interface SharedCategorySummary extends Category {
    permissions: 'view' | 'edit' | 'owner';
}

export interface SharingDashboard {
    shared_categories: SharedCategorySummary[];
    shared_invoices: Invoice[];
    balances: CategoryBalance[];
}