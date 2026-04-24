package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"cogni-cash/internal/domain/entity"
)

// SubscriptionRepository implements port.SubscriptionRepository using pgx.
type SubscriptionRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger
}

// NewSubscriptionRepository creates a new SubscriptionRepository.
func NewSubscriptionRepository(pool *pgxpool.Pool, logger *slog.Logger) *SubscriptionRepository {
	if logger == nil {
		logger = slog.Default()
	}
	return &SubscriptionRepository{pool: pool, Logger: logger}
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Subscription, error) {
	var s entity.Subscription
	err := r.pool.QueryRow(ctx, `
		SELECT 
			id, user_id, merchant_name, amount, currency, billing_cycle, billing_interval,
			category_id, customer_number, contact_email, contact_phone, contact_website,
			support_url, cancellation_url, status, notice_period_days, contract_end_date,
			is_trial, payment_method, last_occurrence, next_occurrence, notes,
			matching_hashes, ignored_hashes, linked_mandates, linked_ibans,
			created_at, updated_at, bank_account_id
		FROM subscriptions
		WHERE id = $1 AND (user_id = $2 
		   OR bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $2)
		   OR category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $2))`, id, userID).
		Scan(
			&s.ID, &s.UserID, &s.MerchantName, &s.Amount, &s.Currency, &s.BillingCycle, &s.BillingInterval,
			&s.CategoryID, &s.CustomerNumber, &s.ContactEmail, &s.ContactPhone, &s.ContactWebsite,
			&s.SupportURL, &s.CancellationURL, &s.Status, &s.NoticePeriodDays, &s.ContractEndDate,
			&s.IsTrial, &s.PaymentMethod, &s.LastOccurrence, &s.NextOccurrence, &s.Notes,
			&s.MatchingHashes, &s.IgnoredHashes, &s.LinkedMandates, &s.LinkedIbans,
			&s.CreatedAt, &s.UpdatedAt, &s.BankAccountID,
		)
	if err != nil {
		return entity.Subscription{}, fmt.Errorf("subscription repo: get by id: %w", err)
	}
	return s, nil
}

func (r *SubscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			id, user_id, merchant_name, amount, currency, billing_cycle, billing_interval,
			category_id, customer_number, contact_email, contact_phone, contact_website,
			support_url, cancellation_url, status, notice_period_days, contract_end_date,
			is_trial, payment_method, last_occurrence, next_occurrence, notes,
			matching_hashes, ignored_hashes, linked_mandates, linked_ibans,
			created_at, updated_at, bank_account_id
		FROM subscriptions
		WHERE user_id = $1 
		   OR bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $1)
		   OR category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $1)
		ORDER BY next_occurrence ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("subscription repo: find by user id: %w", err)
	}
	defer rows.Close()

	var subs []entity.Subscription
	for rows.Next() {
		var s entity.Subscription
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.MerchantName, &s.Amount, &s.Currency, &s.BillingCycle, &s.BillingInterval,
			&s.CategoryID, &s.CustomerNumber, &s.ContactEmail, &s.ContactPhone, &s.ContactWebsite,
			&s.SupportURL, &s.CancellationURL, &s.Status, &s.NoticePeriodDays, &s.ContractEndDate,
			&s.IsTrial, &s.PaymentMethod, &s.LastOccurrence, &s.NextOccurrence, &s.Notes,
			&s.MatchingHashes, &s.IgnoredHashes, &s.LinkedMandates, &s.LinkedIbans,
			&s.CreatedAt, &s.UpdatedAt, &s.BankAccountID,
		); err != nil {
			return nil, fmt.Errorf("subscription repo: scan: %w", err)
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, s entity.Subscription) (entity.Subscription, error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}

	if s.MatchingHashes == nil {
		s.MatchingHashes = []string{}
	}
	if s.IgnoredHashes == nil {
		s.IgnoredHashes = []string{}
	}
	if s.LinkedMandates == nil {
		s.LinkedMandates = []string{}
	}
	if s.LinkedIbans == nil {
		s.LinkedIbans = []string{}
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO subscriptions (
			id, user_id, merchant_name, amount, currency, billing_cycle, billing_interval,
			category_id, customer_number, contact_email, contact_phone, contact_website,
			support_url, cancellation_url, status, notice_period_days, contract_end_date,
			is_trial, payment_method, last_occurrence, next_occurrence, notes,
			matching_hashes, ignored_hashes, linked_mandates, linked_ibans,
			created_at, updated_at, bank_account_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
		) RETURNING created_at, updated_at`,
		s.ID, s.UserID, s.MerchantName, s.Amount, s.Currency, s.BillingCycle, s.BillingInterval,
		s.CategoryID, s.CustomerNumber, s.ContactEmail, s.ContactPhone, s.ContactWebsite,
		s.SupportURL, s.CancellationURL, s.Status, s.NoticePeriodDays, s.ContractEndDate,
		s.IsTrial, s.PaymentMethod, s.LastOccurrence, s.NextOccurrence, s.Notes,
		s.MatchingHashes, s.IgnoredHashes, s.LinkedMandates, s.LinkedIbans,
		s.CreatedAt, s.UpdatedAt, s.BankAccountID,
	).Scan(&s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		return entity.Subscription{}, fmt.Errorf("subscription repo: create: %w", err)
	}
	return s, nil
}

func (r *SubscriptionRepository) CreateWithBackfill(ctx context.Context, sub entity.Subscription, matchingHashes []string) (entity.Subscription, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return entity.Subscription{}, fmt.Errorf("subscription repo: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	now := time.Now()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	sub.MatchingHashes = matchingHashes
	if sub.MatchingHashes == nil {
		sub.MatchingHashes = []string{}
	}
	if sub.IgnoredHashes == nil {
		sub.IgnoredHashes = []string{}
	}
	if sub.LinkedMandates == nil {
		sub.LinkedMandates = []string{}
	}
	if sub.LinkedIbans == nil {
		sub.LinkedIbans = []string{}
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO subscriptions (
			id, user_id, merchant_name, amount, currency, billing_cycle, billing_interval,
			category_id, customer_number, contact_email, contact_phone, contact_website,
			support_url, cancellation_url, status, notice_period_days, contract_end_date,
			is_trial, payment_method, last_occurrence, next_occurrence, notes,
			matching_hashes, ignored_hashes, linked_mandates, linked_ibans,
			created_at, updated_at, bank_account_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
		) RETURNING created_at, updated_at`,
		sub.ID, sub.UserID, sub.MerchantName, sub.Amount, sub.Currency, sub.BillingCycle, sub.BillingInterval,
		sub.CategoryID, sub.CustomerNumber, sub.ContactEmail, sub.ContactPhone, sub.ContactWebsite,
		sub.SupportURL, sub.CancellationURL, sub.Status, sub.NoticePeriodDays, sub.ContractEndDate,
		sub.IsTrial, sub.PaymentMethod, sub.LastOccurrence, sub.NextOccurrence, sub.Notes,
		sub.MatchingHashes, sub.IgnoredHashes, sub.LinkedMandates, sub.LinkedIbans,
		sub.CreatedAt, sub.UpdatedAt, sub.BankAccountID,
	).Scan(&sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return entity.Subscription{}, fmt.Errorf("subscription repo: insert: %w", err)
	}

	if len(matchingHashes) > 0 {
		_, err = tx.Exec(ctx, `
			UPDATE transactions
			SET subscription_id = $1
			WHERE content_hash = ANY($2) AND user_id = $3`,
			sub.ID, matchingHashes, sub.UserID,
		)
		if err != nil {
			return entity.Subscription{}, fmt.Errorf("subscription repo: backfill: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return entity.Subscription{}, fmt.Errorf("subscription repo: commit: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, s entity.Subscription) (entity.Subscription, error) {
	s.UpdatedAt = time.Now()
	if s.MatchingHashes == nil {
		s.MatchingHashes = []string{}
	}
	if s.IgnoredHashes == nil {
		s.IgnoredHashes = []string{}
	}
	if s.LinkedMandates == nil {
		s.LinkedMandates = []string{}
	}
	if s.LinkedIbans == nil {
		s.LinkedIbans = []string{}
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE subscriptions SET
			merchant_name = $1, amount = $2, currency = $3, billing_cycle = $4, billing_interval = $5,
			category_id = $6, customer_number = $7, contact_email = $8, contact_phone = $9, contact_website = $10,
			support_url = $11, cancellation_url = $12, status = $13, notice_period_days = $14, contract_end_date = $15,
			is_trial = $16, payment_method = $17, last_occurrence = $18, next_occurrence = $19, notes = $20,
			matching_hashes = $21, ignored_hashes = $22, linked_mandates = $23, linked_ibans = $24,
			updated_at = $25, bank_account_id = $26
		WHERE id = $27 AND user_id = $28`,
		s.MerchantName, s.Amount, s.Currency, s.BillingCycle, s.BillingInterval,
		s.CategoryID, s.CustomerNumber, s.ContactEmail, s.ContactPhone, s.ContactWebsite,
		s.SupportURL, s.CancellationURL, s.Status, s.NoticePeriodDays, s.ContractEndDate,
		s.IsTrial, s.PaymentMethod, s.LastOccurrence, s.NextOccurrence, s.Notes,
		s.MatchingHashes, s.IgnoredHashes, s.LinkedMandates, s.LinkedIbans,
		s.UpdatedAt, s.BankAccountID, s.ID, s.UserID,
	)
	if err != nil {
		return entity.Subscription{}, fmt.Errorf("subscription repo: update: %w", err)
	}
	return s, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("subscription repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscription repo: not found: %s", id)
	}
	return nil
}

func (r *SubscriptionRepository) LogEvent(ctx context.Context, event entity.SubscriptionEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscription_events (id, subscription_id, user_id, event_type, title, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		event.ID, event.SubscriptionID, event.UserID, event.EventType, event.Title, event.Content, event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("subscription repo: log event: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetEvents(ctx context.Context, subID uuid.UUID, userID uuid.UUID) ([]entity.SubscriptionEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, subscription_id, user_id, event_type, title, content, created_at
		FROM subscription_events
		WHERE subscription_id = $1 AND user_id = $2
		ORDER BY created_at DESC`, subID, userID)
	if err != nil {
		return nil, fmt.Errorf("subscription repo: get events: %w", err)
	}
	defer rows.Close()

	var events []entity.SubscriptionEvent
	for rows.Next() {
		var e entity.SubscriptionEvent
		if err := rows.Scan(&e.ID, &e.SubscriptionID, &e.UserID, &e.EventType, &e.Title, &e.Content, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("subscription repo: scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *SubscriptionRepository) SetDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string, status entity.DiscoveryFeedbackStatus, source string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscription_discovery_feedback (user_id, merchant_name, status, source, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, merchant_name) 
		DO UPDATE SET status = EXCLUDED.status, source = EXCLUDED.source, updated_at = NOW()`,
		userID, merchantName, string(status), source)
	if err != nil {
		return fmt.Errorf("subscription repo: set feedback: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetDiscoveryFeedback(ctx context.Context, userID uuid.UUID) ([]entity.DiscoveryFeedback, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, merchant_name, status, source, created_at, updated_at
		FROM subscription_discovery_feedback
		WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("subscription repo: get feedback: %w", err)
	}
	defer rows.Close()

	var feedback []entity.DiscoveryFeedback
	for rows.Next() {
		var f entity.DiscoveryFeedback
		var status string
		if err := rows.Scan(&f.UserID, &f.MerchantName, &status, &f.Source, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("subscription repo: scan feedback: %w", err)
		}
		f.Status = entity.DiscoveryFeedbackStatus(status)
		feedback = append(feedback, f)
	}
	return feedback, nil
}

func (r *SubscriptionRepository) DeleteDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM subscription_discovery_feedback
		WHERE user_id = $1 AND merchant_name = $2`, userID, merchantName)
	if err != nil {
		return fmt.Errorf("subscription repo: delete feedback: %w", err)
	}
	return nil
}
