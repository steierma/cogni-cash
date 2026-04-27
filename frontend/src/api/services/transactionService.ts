import { api } from '../client';
import type { Transaction, CashFlowForecast, TransactionAnalytics, JobState } from "../types/transaction";
import type { AxiosResponse } from 'axios';

export const transactionService = {
    fetchTransactions: (statementId?: string, hideReconciled: boolean = true, categoryId?: string, reviewed?: boolean, search?: string, includePredictions: boolean = false, includeShared: boolean = false, subscriptionId?: string): Promise<Transaction[]> =>
        api.get<Transaction[]>('transactions/', {
            params: {
                statement_id: statementId || undefined, hide_reconciled: hideReconciled,
                category_id: categoryId || undefined, reviewed: reviewed,
                search: search || undefined, include_predictions: includePredictions,
                include_shared: includeShared,
                subscription_id: subscriptionId || undefined
            }
        }).then((r: AxiosResponse<Transaction[]>) => r.data ?? []),

    fetchForecast: (fromDate?: string, toDate?: string): Promise<CashFlowForecast> =>
        api.get<CashFlowForecast>('transactions/forecast/', { params: { from: fromDate, to: toDate } }).then(r => r.data),

    fetchAnalytics: (hideReconciled: boolean = true): Promise<TransactionAnalytics> =>
        api.get<TransactionAnalytics>('transactions/analytics/', { params: { hide_reconciled: hideReconciled } }).then(r => r.data),

    updateCategory: (contentHash: string, categoryId: string): Promise<void> =>
        api.patch(`transactions/${contentHash}/category/`, { category_id: categoryId }).then(() => undefined),

    updateCategoryBulk: (hashes: string[], categoryId: string): Promise<void> =>
        api.patch('transactions/bulk-category/', { hashes, category_id: categoryId }).then(() => undefined),

    markReviewed: (contentHash: string): Promise<void> =>
        api.patch(`transactions/${contentHash}/review/`).then(() => undefined),

    markReviewedBulk: (hashes: string[]): Promise<void> =>
        api.patch('transactions/bulk-review/', { hashes }).then(() => undefined),


    startAutoCategorize: (): Promise<{ message: string }> =>
        api.post<{ message: string }>('transactions/auto-categorize/start/').then(r => r.data),

    getAutoCategorizeStatus: (): Promise<JobState> =>
        api.get<JobState>('transactions/auto-categorize/status/').then(r => r.data),

    cancelAutoCategorize: (): Promise<{ message: string }> =>
        api.post<{ message: string }>('transactions/auto-categorize/cancel/').then(r => r.data)
};