package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"
)

// --- Mocks ---

type mockPayslipRepoForPayslipSvc struct {
	exists bool
	saved  bool
	err    error
}

func (m *mockPayslipRepoForPayslipSvc) Save(_ context.Context, p *entity.Payslip) error {
	if m.err != nil {
		return m.err
	}
	m.saved = true
	p.ID = "generated-uuid"
	return nil
}

func (m *mockPayslipRepoForPayslipSvc) ExistsByHash(_ context.Context, _ string) (bool, error) {
	return m.exists, nil
}

func (m *mockPayslipRepoForPayslipSvc) ExistsByOriginalFileName(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockPayslipRepoForPayslipSvc) Update(_ context.Context, p *entity.Payslip) error {
	return m.err
}
func (m *mockPayslipRepoForPayslipSvc) Delete(_ context.Context, _ string) error { return m.err }
func (m *mockPayslipRepoForPayslipSvc) FindAll(_ context.Context) ([]entity.Payslip, error) {
	return nil, nil
}
func (m *mockPayslipRepoForPayslipSvc) FindByID(_ context.Context, _ string) (entity.Payslip, error) {
	return entity.Payslip{}, nil
}
func (m *mockPayslipRepoForPayslipSvc) GetOriginalFile(_ context.Context, _ string) ([]byte, string, string, error) {
	return nil, "", "", nil
}

type mockPayslipParserForPayslipSvc struct {
	payslip entity.Payslip
	err     error
	called  bool
}

func (m *mockPayslipParserForPayslipSvc) Parse(_ context.Context, _ string) (entity.Payslip, error) {
	m.called = true
	return m.payslip, m.err
}

// --- Import Tests ---

func TestPayslipService_Import(t *testing.T) {
	t.Run("Successful Static Import (No Fallback)", func(t *testing.T) {
		repo := &mockPayslipRepoForPayslipSvc{exists: false}
		staticParser := &mockPayslipParserForPayslipSvc{
			payslip: entity.Payslip{
				EmployeeName:   "John Doe",
				GrossPay:       4500.50,
				PeriodMonthNum: 3,
			},
		}
		aiParser := &mockPayslipParserForPayslipSvc{}

		svc := service.NewPayslipService(repo, staticParser, aiParser, nopLogger())

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
		repo := &mockPayslipRepoForPayslipSvc{exists: false}
		staticParser := &mockPayslipParserForPayslipSvc{
			payslip: entity.Payslip{
				EmployeeName:   "John Doe",
				GrossPay:       0,
				PeriodMonthNum: 3,
			},
		}
		aiParser := &mockPayslipParserForPayslipSvc{
			payslip: entity.Payslip{
				EmployeeName:   "John Doe",
				GrossPay:       4500.50,
				PeriodMonthNum: 3,
			},
		}

		svc := service.NewPayslipService(repo, staticParser, aiParser, nopLogger())

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
		repo := &mockPayslipRepoForPayslipSvc{exists: false}
		staticParser := &mockPayslipParserForPayslipSvc{
			payslip: entity.Payslip{
				EmployeeName:   "Static Name",
				GrossPay:       1000.00,
				PeriodMonthNum: 1,
			},
		}
		aiParser := &mockPayslipParserForPayslipSvc{
			payslip: entity.Payslip{
				EmployeeName:   "AI Name",
				GrossPay:       4500.50,
				PeriodMonthNum: 3,
			},
		}

		svc := service.NewPayslipService(repo, staticParser, aiParser, nopLogger())

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
		repo := &mockPayslipRepoForPayslipSvc{exists: false}
		staticParser := &mockPayslipParserForPayslipSvc{
			payslip: entity.Payslip{
				EmployeeName:   "Wrong Name Parser Output",
				GrossPay:       4500.50,
				PeriodMonthNum: 3,
			},
		}
		aiParser := &mockPayslipParserForPayslipSvc{}

		svc := service.NewPayslipService(repo, staticParser, aiParser, nopLogger())

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
		repo := &mockPayslipRepoForPayslipSvc{exists: false}
		staticParser := &mockPayslipParserForPayslipSvc{err: errors.New("static fail")}
		aiParser := &mockPayslipParserForPayslipSvc{err: errors.New("ai fail")}

		svc := service.NewPayslipService(repo, staticParser, aiParser, nopLogger())

		_, err := svc.Import(context.Background(), "test.pdf", "test.pdf", "application/pdf", []byte("valid-data"), nil, false)

		if err == nil {
			t.Fatal("Expected error when both parsers fail, got nil")
		}
	})

	t.Run("Duplicate Hash Rejection", func(t *testing.T) {
		repo := &mockPayslipRepoForPayslipSvc{exists: true}
		svc := service.NewPayslipService(repo, &mockPayslipParserForPayslipSvc{}, &mockPayslipParserForPayslipSvc{}, nopLogger())

		_, err := svc.Import(context.Background(), "test.pdf", "test.pdf", "application/pdf", []byte("duplicate-data"), nil, false)

		if err == nil {
			t.Fatal("Expected error for duplicate hash, got nil")
		}
	})
}

// --- Update Tests ---

func TestPayslipService_Update_Success(t *testing.T) {
	repo := &mockPayslipRepoForPayslipSvc{}
	svc := service.NewPayslipService(repo, nil, nil, nopLogger())

	err := svc.Update(context.Background(), &entity.Payslip{
		ID:             "existing-id",
		PeriodMonthNum: 3,
		GrossPay:       5000,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPayslipService_Update_WithNewFile_DuplicateHash(t *testing.T) {
	repo := &mockPayslipRepoForPayslipSvc{exists: true}
	svc := service.NewPayslipService(repo, nil, nil, nopLogger())

	err := svc.Update(context.Background(), &entity.Payslip{
		ID:                  "existing-id",
		OriginalFileContent: []byte("file-content"),
	})
	if !errors.Is(err, service.ErrPayslipDuplicate) {
		t.Errorf("expected ErrPayslipDuplicate, got: %v", err)
	}
}

// --- Delete Tests ---

func TestPayslipService_Delete_Success(t *testing.T) {
	repo := &mockPayslipRepoForPayslipSvc{}
	svc := service.NewPayslipService(repo, nil, nil, nopLogger())

	err := svc.Delete(context.Background(), "some-id")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPayslipService_Delete_RepoError(t *testing.T) {
	repo := &mockPayslipRepoForPayslipSvc{err: errors.New("db error")}
	svc := service.NewPayslipService(repo, nil, nil, nopLogger())

	err := svc.Delete(context.Background(), "some-id")
	if err == nil {
		t.Error("expected error when repo returns error")
	}
}

// --- ImportFromJSONFile Tests ---

func TestPayslipService_ImportFromJSONFile(t *testing.T) {
	ctx := context.Background()

	writeJSONFile := func(t *testing.T, entries []service.JSONPayslipEntry) string {
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
		svc := service.NewPayslipService(&mockPayslipRepoForPayslipSvc{}, &mockPayslipParserForPayslipSvc{}, &mockPayslipParserForPayslipSvc{}, nopLogger())
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
		entries := []service.JSONPayslipEntry{
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

		repo := &mockPayslipRepoForPayslipSvc{}
		svc := service.NewPayslipService(repo, &mockPayslipParserForPayslipSvc{}, &mockPayslipParserForPayslipSvc{}, nopLogger())

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
		entries := []service.JSONPayslipEntry{
			{
				PeriodMonthNum: 5, PeriodYear: 2099, EmployeeName: "Test User",
				GrossPay: 5200, NetPay: 3600, PayoutAmount: 3500,
				OriginalFileName: "Entgeltnachweis_2099_05_31.pdf",
			},
		}
		path := writeJSONFile(t, entries)

		repo := &mockPayslipRepoWithFileNameCheckForPayslipSvc{alreadyExists: true}
		svc := service.NewPayslipService(repo, &mockPayslipParserForPayslipSvc{}, &mockPayslipParserForPayslipSvc{}, nopLogger())

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
		entries := []service.JSONPayslipEntry{
			{PeriodMonthNum: 6, PeriodYear: 2099, EmployeeName: "Bad Entry", GrossPay: 1},
		}
		path := writeJSONFile(t, entries)
		svc := service.NewPayslipService(&mockPayslipRepoForPayslipSvc{}, &mockPayslipParserForPayslipSvc{}, &mockPayslipParserForPayslipSvc{}, nopLogger())

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

type mockPayslipRepoWithFileNameCheckForPayslipSvc struct {
	mockPayslipRepoForPayslipSvc
	alreadyExists bool
}

func (m *mockPayslipRepoWithFileNameCheckForPayslipSvc) ExistsByOriginalFileName(_ context.Context, _ string) (bool, error) {
	return m.alreadyExists, nil
}
