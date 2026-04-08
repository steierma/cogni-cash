package port

import "context"

type JobState struct {
	IsRunning bool                     `json:"is_running"`
	Processed int                      `json:"processed"`
	Total     int                      `json:"total"`
	Status    string                   `json:"status"` // "idle", "running", "cancelled", "completed", "error"
	Results   []CategorizedTransaction `json:"results"`
}

type CategorizationUseCase interface {
	StartAutoCategorizeAsync(ctx context.Context, batchSize int) error // Added batchSize
	GetJobStatus() JobState
	CancelJob()
}

type JobTracker interface {
	Start(total int, cancel context.CancelFunc) error
	AddResults(count int, results []CategorizedTransaction)
	Finish(status string)
	Cancel()
	GetState() JobState
}
