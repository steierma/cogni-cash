package memory

import (
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"context"
	"sync"

	"github.com/google/uuid"
)

var _ port.ForecastingRepository = (*ForecastingRepository)(nil)

type ForecastingRepository struct {
	mu                sync.RWMutex
	exclusions        map[uuid.UUID][]entity.ExcludedForecast
	patternExclusions map[uuid.UUID][]entity.PatternExclusion
}

func NewForecastingRepository() *ForecastingRepository {
	return &ForecastingRepository{
		exclusions:        make(map[uuid.UUID][]entity.ExcludedForecast),
		patternExclusions: make(map[uuid.UUID][]entity.PatternExclusion),
	}
}

func (r *ForecastingRepository) SaveExclusion(ctx context.Context, exclusion entity.ExcludedForecast) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if exclusion.ID == uuid.Nil {
		exclusion.ID = uuid.New()
	}

	userExclusions := r.exclusions[exclusion.UserID]
	// Check for existing
	for _, e := range userExclusions {
		if e.ForecastID == exclusion.ForecastID {
			return nil // Duplicate, ignore
		}
	}

	r.exclusions[exclusion.UserID] = append(userExclusions, exclusion)
	return nil
}

func (r *ForecastingRepository) FindExclusions(ctx context.Context, userID uuid.UUID) ([]entity.ExcludedForecast, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Return a copy to avoid race conditions if the caller modifies the slice
	src := r.exclusions[userID]
	dst := make([]entity.ExcludedForecast, len(src))
	copy(dst, src)
	return dst, nil
}

func (r *ForecastingRepository) DeleteExclusion(ctx context.Context, userID uuid.UUID, forecastID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userExclusions := r.exclusions[userID]
	for i, e := range userExclusions {
		if e.ForecastID == forecastID {
			r.exclusions[userID] = append(userExclusions[:i], userExclusions[i+1:]...)
			break
		}
	}
	return nil
}

func (r *ForecastingRepository) SavePatternExclusion(ctx context.Context, exclusion entity.PatternExclusion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if exclusion.ID == uuid.Nil {
		exclusion.ID = uuid.New()
	}

	userExclusions := r.patternExclusions[exclusion.UserID]
	// Check for existing
	for _, e := range userExclusions {
		if e.MatchTerm == exclusion.MatchTerm {
			return nil // Duplicate, ignore
		}
	}

	r.patternExclusions[exclusion.UserID] = append(userExclusions, exclusion)
	return nil
}

func (r *ForecastingRepository) FindPatternExclusions(ctx context.Context, userID uuid.UUID) ([]entity.PatternExclusion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	src := r.patternExclusions[userID]
	dst := make([]entity.PatternExclusion, len(src))
	copy(dst, src)
	return dst, nil
}

func (r *ForecastingRepository) DeletePatternExclusion(ctx context.Context, userID uuid.UUID, matchTerm string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userExclusions := r.patternExclusions[userID]
	for i, e := range userExclusions {
		if e.MatchTerm == matchTerm {
			r.patternExclusions[userID] = append(userExclusions[:i], userExclusions[i+1:]...)
			break
		}
	}
	return nil
}
