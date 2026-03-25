package service

import (
	"context"
	"errors"
	"sync"

	"cogni-cash/internal/domain/port"
)

// Ensure JobManager implements the port at compile time
var _ port.JobTracker = (*JobManager)(nil)

// ErrJobAlreadyRunning is returned when trying to start a new job while one is in progress.
var ErrJobAlreadyRunning = errors.New("a batch categorization job is already running")

// ErrNothingToCategorize is returned when there are no uncategorized transactions.
var ErrNothingToCategorize = errors.New("no uncategorized transactions found")

type JobManager struct {
	mu         sync.RWMutex
	state      port.JobState
	cancelFunc context.CancelFunc
}

func NewJobManager() *JobManager {
	return &JobManager{
		state: port.JobState{
			Status:  "idle",
			Results: make([]port.CategorizedTransaction, 0),
		},
	}
}

func (jm *JobManager) Start(total int, cancel context.CancelFunc) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if jm.state.IsRunning {
		return ErrJobAlreadyRunning
	}

	jm.state.IsRunning = true
	jm.state.Total = total
	jm.state.Processed = 0
	jm.state.Status = "running"
	jm.state.Results = make([]port.CategorizedTransaction, 0)
	jm.cancelFunc = cancel

	return nil
}

func (jm *JobManager) AddResults(count int, results []port.CategorizedTransaction) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.state.Processed += count
	jm.state.Results = append(jm.state.Results, results...)
}

func (jm *JobManager) Finish(status string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.state.IsRunning = false
	jm.state.Status = status
	jm.cancelFunc = nil
}

func (jm *JobManager) Cancel() {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if jm.cancelFunc != nil {
		jm.cancelFunc()
		jm.state.Status = "cancelled"
	}
}

func (jm *JobManager) GetState() port.JobState {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	resCopy := make([]port.CategorizedTransaction, len(jm.state.Results))
	copy(resCopy, jm.state.Results)

	return port.JobState{
		IsRunning: jm.state.IsRunning,
		Processed: jm.state.Processed,
		Total:     jm.state.Total,
		Status:    jm.state.Status,
		Results:   resCopy,
	}
}
