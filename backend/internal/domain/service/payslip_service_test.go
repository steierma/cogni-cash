package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"cogni-cash/internal/domain/entity"
)

// --- Mocks ---

type mockPayslipRepo struct {
	exists bool
	saved  bool
	err    error
}

func (m *mockPayslipRepo) Save(ctx context.Context, p *entity.Payslip) error {
	if m.err != nil {
		return m.err
	}
	m.saved = true
	p.ID = "generated-uuid"
	return nil
}

func (m *mockPayslipRepo) ExistsByHash(ctx context.Context, hash string) (bool, error) {
	return m.exists, nil
}

func (m *mockPayslipRepo) ExistsByOriginalFileName(ctx context.Context, originalFileName string) (bool, error) {
	return false, nil
}

func (m *mockPayslipRepo) Update(ctx context.Context, p *entity.Payslip) error   { return nil }
func (m *mockPayslipRepo) Delete(ctx context.Context, id string) error           { return nil }
func (m *mockPayslipRepo) FindAll(ctx context.Context) ([]entity.Payslip, error) { return nil, nil }
func (m *mockPayslipRepo) FindByID(ctx context.Context, id string) (entity.Payslip, error) {
	return entity.Payslip{}, nil
}
func (m *mockPayslipRepo) GetOriginalFile(ctx context.Context, id string) ([]byte, string, string, error) {
	return nil, "", "", nil
}

type mockPayslipParser struct {
	payslip entity.Payslip
	err     error
	called  bool
}

func (m *mockPayslipParser) Parse(ctx context.Context, filePath string) (entity.Payslip, error) {
	m.called = true
	return m.payslip, m.err
}

// --- Tests ---

func TestPayslipService_Import(t *testing.T) {
	nopLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Successful Static Import (No Fallback)", func(t *testing.T) {
		repo := &mockPayslipRepo{exists: false}
		staticParser := &mockPayslipParser{
			payslip: entity.Payslip{
				EmployeeName:   "John Doe",
				GrossPay:       4500.50,
				PeriodMonthNum: 3, // Added so the service sees the extraction as complete
			},
		}
		aiParser := &mockPayslipParser{}

		svc := NewPayslipService(repo, staticParser, aiParser, nopLogger)

		res, err := svc.Import(context.Background(), "test.pdf", "test.pdf", "application/pdf", []byte("valid-data"), nil, false)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if aiParser.called {
			t.Error("AI Fallback was triggered even though static parser data was complete")
		}
		if res.ContentHash == "" {
			t.Error("Content hash was not generated")
		}
		if !repo.saved {
			t.Error("Repo.Save was not called")
		}
	})

	t.Run("Trigger AI Fallback when Static Data is Incomplete", func(t *testing.T) {
		repo := &mockPayslipRepo{exists: false}
		staticParser := &mockPayslipParser{
			payslip: entity.Payslip{
				EmployeeName:   "John Doe",
				GrossPay:       0, // Missing gross triggers fallback
				PeriodMonthNum: 3,
			},
		}
		aiParser := &mockPayslipParser{
			payslip: entity.Payslip{
				EmployeeName:   "John Doe",
				GrossPay:       4500.50,
				PeriodMonthNum: 3,
			},
		}

		svc := NewPayslipService(repo, staticParser, aiParser, nopLogger)

		res, err := svc.Import(context.Background(), "test.pdf", "test.pdf", "application/pdf", []byte("valid-data"), nil, false)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !aiParser.called {
			t.Error("AI Fallback should have been triggered for incomplete static data")
		}
		if res.GrossPay != 4500.50 {
			t.Errorf("Expected AI data 4500.50, got %f", res.GrossPay)
		}
	})

	t.Run("Force AI Parsing", func(t *testing.T) {
		repo := &mockPayslipRepo{exists: false}
		staticParser := &mockPayslipParser{
			payslip: entity.Payslip{
				EmployeeName:   "Static Name",
				GrossPay:       1000.00,
				PeriodMonthNum: 1,
			},
		}
		aiParser := &mockPayslipParser{
			payslip: entity.Payslip{
				EmployeeName:   "AI Name",
				GrossPay:       4500.50,
				PeriodMonthNum: 3,
			},
		}

		svc := NewPayslipService(repo, staticParser, aiParser, nopLogger)

		res, err := svc.Import(context.Background(), "test.pdf", "test.pdf", "application/pdf", []byte("valid-data"), nil, true)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if staticParser.called {
			t.Error("Static parser was called even though useAI was true")
		}
		if !aiParser.called {
			t.Error("AI parser was not called even though useAI was true")
		}
		if res.EmployeeName != "AI Name" {
			t.Errorf("Expected AI data 'AI Name', got %s", res.EmployeeName)
		}
	})

	t.Run("Apply Manual Overrides", func(t *testing.T) {
		repo := &mockPayslipRepo{exists: false}
		staticParser := &mockPayslipParser{
			payslip: entity.Payslip{
				EmployeeName:   "Wrong Name Parser Output",
				GrossPay:       4500.50,
				PeriodMonthNum: 3,
			},
		}
		aiParser := &mockPayslipParser{}

		svc := NewPayslipService(repo, staticParser, aiParser, nopLogger)

		overrides := &entity.Payslip{
			EmployeeName:   "Correct Manual Name",
			GrossPay:       5000.00,
			PeriodMonthNum: 4,
		}

		res, err := svc.Import(context.Background(), "test.pdf", "test.pdf", "application/pdf", []byte("valid-data"), overrides, false)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res.EmployeeName != "Correct Manual Name" {
			t.Errorf("Expected EmployeeName to be overridden, got %s", res.EmployeeName)
		}
		if res.GrossPay != 5000.00 {
			t.Errorf("Expected GrossPay to be overridden, got %f", res.GrossPay)
		}
	})

	t.Run("Fail when Both Parsers Fail", func(t *testing.T) {
		repo := &mockPayslipRepo{exists: false}
		staticParser := &mockPayslipParser{err: errors.New("static fail")}
		aiParser := &mockPayslipParser{err: errors.New("ai fail")}

		svc := NewPayslipService(repo, staticParser, aiParser, nopLogger)

		_, err := svc.Import(context.Background(), "test.pdf", "test.pdf", "application/pdf", []byte("valid-data"), nil, false)

		if err == nil {
			t.Fatal("Expected error when both parsers fail, got nil")
		}
	})

	t.Run("Duplicate Hash Rejection", func(t *testing.T) {
		repo := &mockPayslipRepo{exists: true}
		svc := NewPayslipService(repo, &mockPayslipParser{}, &mockPayslipParser{}, nopLogger)

		_, err := svc.Import(context.Background(), "test.pdf", "test.pdf", "application/pdf", []byte("duplicate-data"), nil, false)

		if err == nil {
			t.Fatal("Expected error for duplicate hash, got nil")
		}
	})
}

func TestPayslipService_ImportFromJSONFile(t *testing.T) {
	nopLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	writeJSON := func(t *testing.T, entries []jsonPayslipEntry) string {
		t.Helper()
		f, err := os.CreateTemp(t.TempDir(), "payslips_import_*.json")
		if err != nil {
			t.Fatalf("create temp file: %v", err)
		}
		if err := json.NewEncoder(f).Encode(entries); err != nil {
			t.Fatalf("encode json: %v", err)
		}
		f.Close()
		return f.Name()
	}

	t.Run("Non-existent file is a no-op", func(t *testing.T) {
		svc := NewPayslipService(&mockPayslipRepo{}, &mockPayslipParser{}, &mockPayslipParser{}, nopLogger)
		imported, skipped, errs, fatalErr := svc.ImportFromJSONFile(ctx, "/does/not/exist.json")
		if fatalErr != nil {
			t.Fatalf("unexpected fatal error: %v", fatalErr)
		}
		if imported != 0 || skipped != 0 || len(errs) != 0 {
			t.Errorf("expected all-zero results, got imported=%d skipped=%d errs=%d", imported, skipped, len(errs))
		}
	})

	t.Run("Imports new entries, deletes PDFs, keeps JSON", func(t *testing.T) {
		dir := t.TempDir()
		entries := []jsonPayslipEntry{
			{
				PeriodMonthNum: 3, PeriodYear: 2099, EmployeeName: "Test User",
				GrossPay: 5000, NetPay: 3500, PayoutAmount: 3400,
				OriginalFileName: "Entgeltnachweis_2099_03_31.pdf",
			},
			{
				PeriodMonthNum: 4, PeriodYear: 2099, EmployeeName: "Test User",
				GrossPay: 5100, NetPay: 3550, PayoutAmount: 3450,
				OriginalFileName: "Entgeltnachweis_2099_04_30.pdf",
				Bonuses:          []entity.Bonus{{Description: "Bonus", Amount: 500}},
			},
		}

		jsonPath := filepath.Join(dir, "payslips_import.json")
		f, _ := os.Create(jsonPath)
		json.NewEncoder(f).Encode(entries)
		f.Close()

		for _, e := range entries {
			os.WriteFile(filepath.Join(dir, e.OriginalFileName), []byte("dummy"), 0644)
		}

		repo := &mockPayslipRepo{}
		svc := NewPayslipService(repo, &mockPayslipParser{}, &mockPayslipParser{}, nopLogger)

		imported, skipped, errs, fatalErr := svc.ImportFromJSONFile(ctx, jsonPath)

		if fatalErr != nil {
			t.Fatalf("unexpected fatal error: %v", fatalErr)
		}
		if len(errs) != 0 {
			t.Fatalf("unexpected per-record errors: %v", errs)
		}
		if imported != 2 {
			t.Errorf("expected 2 imported, got %d", imported)
		}
		if skipped != 0 {
			t.Errorf("expected 0 skipped, got %d", skipped)
		}
		if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
			t.Error("expected JSON manifest to be kept after import")
		}
		for _, e := range entries {
			pdfPath := filepath.Join(dir, e.OriginalFileName)
			if _, err := os.Stat(pdfPath); !os.IsNotExist(err) {
				t.Errorf("expected PDF %s to be deleted after import", e.OriginalFileName)
			}
		}
	})

	t.Run("Skips duplicate entries by original_file_name", func(t *testing.T) {
		entries := []jsonPayslipEntry{
			{
				PeriodMonthNum: 5, PeriodYear: 2099, EmployeeName: "Test User",
				GrossPay: 5200, NetPay: 3600, PayoutAmount: 3500,
				OriginalFileName: "Entgeltnachweis_2099_05_31.pdf",
			},
		}
		path := writeJSON(t, entries)

		repo := &mockPayslipRepoWithFileNameCheck{alreadyExists: true}
		svc := NewPayslipService(repo, &mockPayslipParser{}, &mockPayslipParser{}, nopLogger)

		imported, skipped, errs, fatalErr := svc.ImportFromJSONFile(ctx, path)

		if fatalErr != nil {
			t.Fatalf("unexpected fatal error: %v", fatalErr)
		}
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if imported != 0 {
			t.Errorf("expected 0 imported (all dupes), got %d", imported)
		}
		if skipped != 1 {
			t.Errorf("expected 1 skipped, got %d", skipped)
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("expected JSON manifest to be kept after skipped import")
		}
	})

	t.Run("JSON kept and PDF not deleted when save fails", func(t *testing.T) {
		entries := []jsonPayslipEntry{
			{PeriodMonthNum: 6, PeriodYear: 2099, EmployeeName: "Bad Entry", GrossPay: 1},
		}
		path := writeJSON(t, entries)
		svc := NewPayslipService(&mockPayslipRepo{}, &mockPayslipParser{}, &mockPayslipParser{}, nopLogger)

		_, _, errs, fatalErr := svc.ImportFromJSONFile(ctx, path)

		if fatalErr != nil {
			t.Fatalf("unexpected fatal error: %v", fatalErr)
		}
		if len(errs) == 0 {
			t.Error("expected at least one per-record error for missing original_file_name")
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("expected JSON manifest to be kept when errors occurred")
		}
	})
}

type mockPayslipRepoWithFileNameCheck struct {
	mockPayslipRepo
	alreadyExists bool
}

func (m *mockPayslipRepoWithFileNameCheck) ExistsByOriginalFileName(_ context.Context, _ string) (bool, error) {
	return m.alreadyExists, nil
}
