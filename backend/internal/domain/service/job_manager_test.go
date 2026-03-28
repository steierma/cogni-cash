package service_test

import (
	"context"
	"testing"

	"cogni-cash/internal/domain/service"
)

func TestJobManager_Cancel(t *testing.T) {
	jm := service.NewJobManager()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := jm.Start(10, cancel)
	if err != nil {
		t.Fatalf("expected no error starting job, got: %v", err)
	}

	state := jm.GetState()
	if !state.IsRunning {
		t.Error("expected job to be running")
	}

	jm.Cancel()
	state = jm.GetState()
	if state.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got %q", state.Status)
	}

	// Verify the context was cancelled
	select {
	case <-ctx.Done():
		// OK — context was properly cancelled
	default:
		t.Error("expected cancel func to be called and context to be done")
	}
}

func TestJobManager_DoubleStart(t *testing.T) {
	jm := service.NewJobManager()

	_, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	_, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	err := jm.Start(10, cancel1)
	if err != nil {
		t.Fatalf("expected no error on first start, got: %v", err)
	}

	err = jm.Start(5, cancel2)
	if err == nil {
		t.Error("expected error when starting a second job while first is running")
	}
}

func TestJobManager_FinishThenRestart(t *testing.T) {
	jm := service.NewJobManager()

	_, cancel1 := context.WithCancel(context.Background())
	defer cancel1()

	err := jm.Start(10, cancel1)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	jm.Finish("completed")

	state := jm.GetState()
	if state.IsRunning {
		t.Error("expected job to not be running after finish")
	}
	if state.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", state.Status)
	}

	// Should be able to start a new job
	_, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	err = jm.Start(5, cancel2)
	if err != nil {
		t.Fatalf("expected to restart after finish, got: %v", err)
	}
}

func TestJobManager_CancelWhenIdle(t *testing.T) {
	jm := service.NewJobManager()

	// Should not panic when cancelling an idle job
	jm.Cancel()

	state := jm.GetState()
	if state.Status != "idle" {
		t.Errorf("expected status to remain 'idle', got %q", state.Status)
	}
}

func TestTransactionService_CancelJob(t *testing.T) {
	repo := &mockRepo{}
	catRepo := &mockCategoryRepo{}

	svc := service.NewTransactionService(repo, catRepo, nil, nil, setupLogger())

	// Should not panic
	svc.CancelJob()

	status := svc.GetJobStatus()
	if status.Status != "idle" {
		t.Errorf("expected idle status, got %q", status.Status)
	}
}

