import { api } from '../client';
import type { CashFlowForecast, PlannedTransaction, CreatePlannedTransactionRequest, UpdatePlannedTransactionRequest } from "../types/transaction";
export const forecastingService = {
    fetchForecast: (fromDate?: string, toDate?: string): Promise<CashFlowForecast> =>
        api.get<CashFlowForecast>('transactions/forecast/', { params: { from: fromDate, to: toDate } }).then(r => r.data),

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
