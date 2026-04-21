import { api } from '../client';
import type { 
    Subscription, 
    SuggestedSubscription, 
    CancellationLetterResult, 
    SubscriptionEvent,
    DiscoveryFeedback
} from "../types/subscription";

export const subscriptionService = {
    fetchSubscriptions: (): Promise<Subscription[]> =>
        api.get<Subscription[]>('subscriptions/').then(r => r.data ?? []),

    getSubscription: (id: string): Promise<Subscription> =>
        api.get<Subscription>(`subscriptions/${id}/`).then(r => r.data),

    updateSubscription: (id: string, sub: Partial<Subscription>): Promise<Subscription> =>
        api.put<Subscription>(`subscriptions/${id}/`, sub).then(r => r.data),

    fetchSuggestedSubscriptions: (): Promise<SuggestedSubscription[]> =>
        api.get<SuggestedSubscription[]>('subscriptions/suggested/').then(r => r.data ?? []),

    fetchDiscoveryFeedback: (): Promise<DiscoveryFeedback[]> =>
        api.get<DiscoveryFeedback[]>('subscriptions/feedback/').then(r => r.data ?? []),

    approveSubscription: (suggestion: SuggestedSubscription): Promise<Subscription> =>
        api.post<Subscription>('subscriptions/approve/', { suggestion }).then(r => r.data),

    declineSuggestion: (merchantName: string): Promise<void> =>
        api.post('subscriptions/decline/', { merchant_name: merchantName }).then(() => undefined),

    removeDiscoveryFeedback: (merchantName: string): Promise<void> =>
        api.post('subscriptions/remove-feedback/', { merchant_name: merchantName }).then(() => undefined),

    /** @deprecated Use fetchDiscoveryFeedback instead */
    fetchDeclinedSubscriptions: (): Promise<DiscoveryFeedback[]> =>
        api.get<DiscoveryFeedback[]>('subscriptions/feedback/').then(r => r.data ?? []),

    /** @deprecated Use removeDiscoveryFeedback instead */
    undeclineSuggestion: (merchantName: string): Promise<void> =>
        api.post('subscriptions/remove-feedback/', { merchant_name: merchantName }).then(() => undefined),

    deleteSubscription: (id: string): Promise<void> =>
        api.delete(`subscriptions/${id}/`).then(() => undefined),

    enrichSubscription: (id: string): Promise<Subscription> =>
        api.post<Subscription>(`subscriptions/${id}/enrich/`).then(r => r.data),

    previewCancellation: (id: string, lang?: string): Promise<CancellationLetterResult> =>
        api.post<CancellationLetterResult>(`subscriptions/${id}/preview-cancellation/`, null, { params: { lang } }).then(r => r.data),

    cancelSubscription: (id: string, subject: string, body: string): Promise<void> =>
        api.post(`subscriptions/${id}/cancel/`, { subject, body }).then(() => undefined),

    fetchSubscriptionEvents: (id: string): Promise<SubscriptionEvent[]> =>
        api.get<SubscriptionEvent[]>(`subscriptions/${id}/events/`).then(r => r.data ?? []),

    linkTransaction: (subID: string, txnHash: string): Promise<void> =>
        api.post(`subscriptions/${subID}/transactions/${txnHash}/link/`).then(() => undefined),

    unlinkTransaction: (subID: string, txnHash: string): Promise<void> =>
        api.delete(`subscriptions/${subID}/transactions/${txnHash}/unlink/`).then(() => undefined),

    createFromTransaction: (transactionHash: string, billingCycle: string): Promise<Subscription> =>
        api.post<Subscription>('subscriptions/from-transaction/', { 
            transaction_hash: transactionHash, 
            billing_cycle: billingCycle 
        }).then(r => r.data),
};
