export type SubscriptionStatus = 'active' | 'cancellation_pending' | 'cancelled' | 'paused';

export interface Subscription {
    id: string;
    user_id: string;
    merchant_name: string;
    amount: number;
    currency: string;
    billing_cycle: 'monthly' | 'yearly';
    billing_interval: number;
    category_id: string | null;
    customer_number?: string;
    contact_email?: string;
    contact_phone?: string;
    contact_website?: string;
    support_url?: string;
    cancellation_url?: string;
    status: SubscriptionStatus;
    notice_period_days: number;
    contract_end_date?: string;
    is_trial: boolean;
    payment_method?: string;
    last_occurrence?: string;
    next_occurrence?: string;
    notes?: string;
    matching_hashes: string[];
    ignored_hashes: string[];
    created_at: string;
    updated_at: string;
}

export interface BaseTransaction {
    date: string;
    amount: number;
}

export interface SuggestedSubscription {
    merchant_name: string;
    estimated_amount: number;
    currency: string;
    billing_cycle: 'monthly' | 'yearly';
    billing_interval: number;
    last_occurrence: string;
    next_occurrence: string;
    matching_hashes: string[];
    base_transactions: BaseTransaction[];
    category_id: string | null;
}

export interface ApproveSubscriptionRequest {
    suggestion: SuggestedSubscription;
}

export interface CancellationLetterResult {
    subject: string;
    body: string;
}

export interface SubscriptionEvent {
    id: string;
    subscription_id: string;
    user_id: string;
    event_type: string;
    title: string;
    content: string;
    created_at: string;
}

export type DiscoveryFeedbackStatus = 'ALLOWED' | 'DECLINED' | 'AI_REJECTED';

export interface DiscoveryFeedback {
    user_id: string;
    merchant_name: string;
    status: DiscoveryFeedbackStatus;
    source: string;
    created_at: string;
    updated_at: string;
}

/** @deprecated Use DiscoveryFeedback instead */
export type DeclinedSubscription = DiscoveryFeedback;
