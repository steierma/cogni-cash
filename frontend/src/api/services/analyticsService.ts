import { api } from '../client';
import type { TransactionAnalytics } from "../types/transaction";
export const analyticsService = {
    fetchAnalytics: (hideReconciled: boolean = true): Promise<TransactionAnalytics> =>
        api.get<TransactionAnalytics>('transactions/analytics/', { params: { hide_reconciled: hideReconciled } }).then(r => r.data)
};
