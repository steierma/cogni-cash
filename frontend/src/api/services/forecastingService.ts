import { api } from '../client';
import type { CashFlowForecast, PlannedTransaction, CreatePlannedTransactionRequest, UpdatePlannedTransactionRequest } from "../types/transaction";
export const forecastingService = {
    fetchForecast: (fromDate?: string, toDate?: string): Promise<CashFlowForecast> =>
        api.get<CashFlowForecast>('transactions/forecast/', { params: { from: fromDate, to: toDate } }).then(r => r.data),

    excludeProjection: (id: string): Promise<void> =>
        api.post(`transactions/forecast/exclude/${id}/`).then(() => undefined),

    includeProjection: (id: string): Promise<void> =>
        api.post(`transactions/forecast/include/${id}/`).then(() => undefined),

    fetchPatternExclusions: (): Promise<any[]> =>
        api.get<any[]>('transactions/forecast/patterns/exclusions/').then(r => r.data ?? []),

    excludePattern: (matchTerm: string): Promise<void> =>
        api.post('transactions/forecast/patterns/exclude/', { match_term: matchTerm }).then(() => undefined),

    includePattern: (matchTerm: string): Promise<void> =>
        api.post('transactions/forecast/patterns/include/', { match_term: matchTerm }).then(() => undefined),

    // Planned Transactions
    fetchPlannedTransactions: (): Promise<PlannedTransaction[]> =>
        api.get<PlannedTransaction[]>('planned-transactions/').then(r => r.data ?? []),

    createPlannedTransaction: (data: CreatePlannedTransactionRequest): Promise<PlannedTransaction> =>
        api.post<PlannedTransaction>('planned-transactions/', data).then(r => r.data),

    updatePlannedTransaction: (id: string, data: UpdatePlannedTransactionRequest): Promise<PlannedTransaction> =>
        api.put<PlannedTransaction>(`planned-transactions/${id}/`, data).then(r => r.data),

    deletePlannedTransaction: (id: string): Promise<void> =>
        api.delete(`planned-transactions/${id}/`).then(() => undefined)
};
