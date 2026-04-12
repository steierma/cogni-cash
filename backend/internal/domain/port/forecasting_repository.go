package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// ForecastingRepository defines the storage operations for forecast exclusions.
type ForecastingRepository interface {
	SaveExclusion(ctx context.Context, exclusion entity.ExcludedForecast) error
	FindExclusions(ctx context.Context, userID uuid.UUID) ([]entity.ExcludedForecast, error)
	DeleteExclusion(ctx context.Context, userID uuid.UUID, forecastID uuid.UUID) error

	SavePatternExclusion(ctx context.Context, exclusion entity.PatternExclusion) error
	FindPatternExclusions(ctx context.Context, userID uuid.UUID) ([]entity.PatternExclusion, error)
	DeletePatternExclusion(ctx context.Context, userID uuid.UUID, matchTerm string) error
}
