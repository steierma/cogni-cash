package entity

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive              SubscriptionStatus = "active"
	SubscriptionStatusCancellationPending SubscriptionStatus = "cancellation_pending"
	SubscriptionStatusCancelled           SubscriptionStatus = "cancelled"
	SubscriptionStatusIgnored             SubscriptionStatus = "ignored"
)

type Subscription struct {
	ID               uuid.UUID          `json:"id"`
	UserID           uuid.UUID          `json:"user_id"`
	MerchantName     string             `json:"merchant_name"`
	Amount           float64            `json:"amount"`
	Currency         string             `json:"currency"`
	BillingCycle     string             `json:"billing_cycle"`    // "weekly", "monthly", "yearly"
	BillingInterval  int                `json:"billing_interval"` // e.g., 1 for every month, 3 for quarterly
	CategoryID       *uuid.UUID         `json:"category_id,omitempty"`
	CustomerNumber   *string            `json:"customer_number,omitempty"`
	ContactEmail     *string            `json:"contact_email,omitempty"`
	ContactPhone     *string            `json:"contact_phone,omitempty"`
	ContactWebsite   *string            `json:"contact_website,omitempty"`
	SupportURL       *string            `json:"support_url,omitempty"`
	CancellationURL  *string            `json:"cancellation_url,omitempty"`
	Status           SubscriptionStatus `json:"status"`
	NoticePeriodDays *int               `json:"notice_period_days,omitempty"`
	ContractEndDate  *time.Time         `json:"contract_end_date,omitempty"`
	IsTrial          bool               `json:"is_trial"`
	PaymentMethod    *string            `json:"payment_method,omitempty"`
	LastOccurrence   *time.Time         `json:"last_occurrence,omitempty"`
	NextOccurrence   *time.Time         `json:"next_occurrence,omitempty"`
	Notes            *string            `json:"notes,omitempty"`
	MatchingHashes   []string           `json:"matching_hashes"`
	IgnoredHashes    []string           `json:"ignored_hashes"`
	LinkedMandates   []string           `json:"linked_mandates"`
	LinkedIbans      []string           `json:"linked_ibans"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

type SuggestedSubscription struct {
	MerchantName     string            `json:"merchant_name"`
	EstimatedAmount  float64           `json:"estimated_amount"`
	Currency         string            `json:"currency"`
	BillingCycle     string            `json:"billing_cycle"`
	BillingInterval  int               `json:"billing_interval"`
	LastOccurrence   time.Time         `json:"last_occurrence"`
	NextOccurrence   time.Time         `json:"next_occurrence"`
	MatchingHashes   []string          `json:"matching_hashes"`
	BaseTransactions []BaseTransaction `json:"base_transactions"`
	CategoryID       *uuid.UUID        `json:"category_id"`
}

type BaseTransaction struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
}

type SubscriptionEvent struct {
	ID             uuid.UUID `json:"id"`
	SubscriptionID uuid.UUID `json:"subscription_id"`
	UserID         uuid.UUID `json:"user_id"`
	EventType      string    `json:"event_type"` // e.g., "cancellation_sent", "status_changed"
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type DiscoveryFeedbackStatus string

const (
	DiscoveryStatusAllowed    DiscoveryFeedbackStatus = "ALLOWED"
	DiscoveryStatusDeclined   DiscoveryFeedbackStatus = "DECLINED"
	DiscoveryStatusAIRejected DiscoveryFeedbackStatus = "AI_REJECTED"
)

type DiscoveryFeedback struct {
	UserID       uuid.UUID               `json:"user_id"`
	MerchantName string                  `json:"merchant_name"`
	Status       DiscoveryFeedbackStatus `json:"status"`
	Source       string                  `json:"source"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}
