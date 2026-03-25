package postgres

import (
	"context"
	"testing"
)

func TestNewPool(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t) // Instant cleanup!

	pool, err := NewPool(ctx, globalPool.Config().ConnString())
	if err != nil {
		t.Fatalf("expected to successfully connect and ping the database, got: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Errorf("expected ping to succeed, got: %v", err)
	}
}
