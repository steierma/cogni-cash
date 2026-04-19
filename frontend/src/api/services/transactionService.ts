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

    markReviewed: (contentHash: string): Promise<void> =>
        api.patch(`transactions/${contentHash}/review/`).then(() => undefined),

    patchSkipForecasting: (contentHash: string, skip: boolean): Promise<void> =>
        api.patch(`transactions/${contentHash}/skip-forecasting/`, { skip }).then(() => undefined),

    excludeForecastProjection: (id: string): Promise<void> =>
        api.post(`transactions/forecast/exclude/${id}/`).then(() => undefined),

    includeForecastProjection: (id: string): Promise<void> =>
        api.post(`transactions/forecast/include/${id}/`).then(() => undefined),

    fetchPatternExclusions: (): Promise<any[]> =>
        api.get<any[]>('transactions/forecast/patterns/exclusions/').then(r => r.data ?? []),

    excludePattern: (matchTerm: string): Promise<void> =>
        api.post('transactions/forecast/patterns/exclude/', { match_term: matchTerm }).then(() => undefined),

    includePattern: (matchTerm: string): Promise<void> =>
        api.post('transactions/forecast/patterns/include/', { match_term: matchTerm }).then(() => undefined),

    startAutoCategorize: (): Promise<{ message: string }> =>
        api.post<{ message: string }>('transactions/auto-categorize/start/').then(r => r.data),

    getAutoCategorizeStatus: (): Promise<JobState> =>
        api.get<JobState>('transactions/auto-categorize/status/').then(r => r.data),

    cancelAutoCategorize: (): Promise<{ message: string }> =>
        api.post<{ message: string }>('transactions/auto-categorize/cancel/').then(r => r.data)
};