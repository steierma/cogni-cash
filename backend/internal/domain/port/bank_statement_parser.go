package port

import (
	"cogni-cash/internal/domain/entity"
	"context"
)

// BankStatementParser is the output port for structured bank-statement extraction.
// Each bank format (ING, Sparkasse, …) provides its own adapter.
type BankStatementParser interface {
	Parse(ctx context.Context, filePath string) (entity.BankStatement, error)
}
