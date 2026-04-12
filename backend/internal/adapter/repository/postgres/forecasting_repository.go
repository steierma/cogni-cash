package postgres

import (
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ port.ForecastingRepository = (*ForecastingRepository)(nil)

type ForecastingRepository struct {
	pool *pgxpool.Pool
}

func NewForecastingRepository(pool *pgxpool.Pool) *ForecastingRepository {
	return &ForecastingRepository{pool: pool}
}

func (r *ForecastingRepository) SaveExclusion(ctx context.Context, exclusion entity.ExcludedForecast) error {
	if exclusion.ID == uuid.Nil {
		exclusion.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO excluded_forecasts (id, user_id, forecast_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, forecast_id) DO NOTHING`,
		exclusion.ID,
		exclusion.UserID,
		exclusion.ForecastID,
		exclusion.CreatedAt,
	)
	return err
}

func (r *ForecastingRepository) FindExclusions(ctx context.Context, userID uuid.UUID) ([]entity.ExcludedForecast, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, forecast_id, created_at
		FROM excluded_forecasts
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exclusions []entity.ExcludedForecast
	for rows.Next() {
		var e entity.ExcludedForecast
		if err := rows.Scan(&e.ID, &e.UserID, &e.ForecastID, &e.CreatedAt); err != nil {
			return nil, err
		}
		exclusions = append(exclusions, e)
	}
	return exclusions, rows.Err()
}

func (r *ForecastingRepository) DeleteExclusion(ctx context.Context, userID uuid.UUID, forecastID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM excluded_forecasts
		WHERE user_id = $1 AND forecast_id = $2`, userID, forecastID)
	return err
}

func (r *ForecastingRepository) SavePatternExclusion(ctx context.Context, exclusion entity.PatternExclusion) error {
	if exclusion.ID == uuid.Nil {
		exclusion.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO pattern_exclusions (id, user_id, match_term, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, match_term) DO NOTHING`,
		exclusion.ID,
		exclusion.UserID,
		exclusion.MatchTerm,
		exclusion.CreatedAt,
	)
	return err
}

func (r *ForecastingRepository) FindPatternExclusions(ctx context.Context, userID uuid.UUID) ([]entity.PatternExclusion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, match_term, created_at
		FROM pattern_exclusions
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exclusions []entity.PatternExclusion
	for rows.Next() {
		var e entity.PatternExclusion
		if err := rows.Scan(&e.ID, &e.UserID, &e.MatchTerm, &e.CreatedAt); err != nil {
			return nil, err
		}
		exclusions = append(exclusions, e)
	}
	return exclusions, rows.Err()
}

func (r *ForecastingRepository) DeletePatternExclusion(ctx context.Context, userID uuid.UUID, matchTerm string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM pattern_exclusions
		WHERE user_id = $1 AND match_term = $2`, userID, matchTerm)
	return err
}
