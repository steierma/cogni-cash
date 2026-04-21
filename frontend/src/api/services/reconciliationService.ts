import { api } from '../client';
import type { Reconciliation, ReconciliationPairSuggestion } from "../types/transaction";
import type { AxiosResponse } from 'axios';

export const reconciliationService = {
    fetchReconciliations: (): Promise<Reconciliation[]> =>
        api.get<Reconciliation[]>('reconciliations/').then((r: AxiosResponse<Reconciliation[]>) => r.data ?? []),

    fetchSuggestions: (windowDays: number = 7): Promise<ReconciliationPairSuggestion[]> =>
        api.get<ReconciliationPairSuggestion[]>('reconciliations/suggestions/', { params: { window: windowDays } }).then((r: AxiosResponse<ReconciliationPairSuggestion[]>) => r.data ?? []),

    create: (settlementHash: string, targetHash: string): Promise<void> =>
        api.post('reconciliations/', { settlement_tx_hash: settlementHash, target_tx_hash: targetHash }).then(() => undefined),

    delete: (id: string): Promise<void> =>
        api.delete(`reconciliations/${id}/`).then(() => undefined)
};
