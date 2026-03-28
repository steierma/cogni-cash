package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
)

// ErrPayslipDuplicate is returned when a payslip with the same content hash already exists.
var ErrPayslipDuplicate = entity.ErrPayslipDuplicate

type PayslipService struct {
	repo         port.PayslipRepository
	staticParser port.PayslipParser
	aiParser     port.PayslipParser
	logger       *slog.Logger
}

func NewPayslipService(repo port.PayslipRepository, staticParser port.PayslipParser, aiParser port.PayslipParser, logger *slog.Logger) *PayslipService {
	return &PayslipService{
		repo:         repo,
		staticParser: staticParser,
		aiParser:     aiParser,
		logger:       logger,
	}
}

func (s *PayslipService) Import(ctx context.Context, filePath, fileName, mimeType string, fileBytes []byte, overrides *entity.Payslip, useAI bool) (*entity.Payslip, error) {
	hashFunc := sha256.New()
	hashFunc.Write(fileBytes)
	contentHash := hex.EncodeToString(hashFunc.Sum(nil))

	exists, err := s.repo.ExistsByHash(ctx, contentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check hash: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: %s", ErrPayslipDuplicate, contentHash)
	}

	var payslip entity.Payslip
	var parseErr error

	if useAI {
		s.logger.Info("Force AI Parsing requested. Bypassing static parser.")
		payslip, parseErr = s.aiParser.Parse(ctx, filePath)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse payslip with forced AI: %w", parseErr)
		}
	} else {
		payslip, parseErr = s.staticParser.Parse(ctx, filePath)

		if parseErr != nil || payslip.PeriodMonthNum == 0 || payslip.EmployeeName == "" || payslip.GrossPay == 0 {

			canSkipAI := overrides != nil && overrides.PeriodMonthNum != 0 && overrides.EmployeeName != "" && overrides.GrossPay != 0

			if !canSkipAI {
				s.logger.Warn("Static parser failed or returned incomplete data. Triggering AI fallback.", "file", fileName, "static_error", parseErr)

				var aiErr error
				aiPayslip, aiErr := s.aiParser.Parse(ctx, filePath)
				if aiErr != nil {
					return nil, fmt.Errorf("failed to parse payslip with AI fallback: %w", aiErr)
				}
				payslip = aiPayslip
			} else {
				s.logger.Info("Static parser failed, but sufficient manual overrides provided. Skipping AI fallback.")
			}
		}
	}

	if overrides != nil {
		if overrides.PeriodMonthNum != 0 {
			payslip.PeriodMonthNum = overrides.PeriodMonthNum
		}
		if overrides.PeriodYear != 0 {
			payslip.PeriodYear = overrides.PeriodYear
		}
		if overrides.EmployeeName != "" {
			payslip.EmployeeName = overrides.EmployeeName
		}
		if overrides.TaxClass != "" {
			payslip.TaxClass = overrides.TaxClass
		}
		if overrides.TaxID != "" {
			payslip.TaxID = overrides.TaxID
		}
		if overrides.GrossPay != 0 {
			payslip.GrossPay = overrides.GrossPay
		}
		if overrides.NetPay != 0 {
			payslip.NetPay = overrides.NetPay
		}
		if overrides.PayoutAmount != 0 {
			payslip.PayoutAmount = overrides.PayoutAmount
		}
		if overrides.CustomDeductions != 0 {
			payslip.CustomDeductions = overrides.CustomDeductions
		}
		if len(overrides.Bonuses) > 0 {
			payslip.Bonuses = overrides.Bonuses
		}
	}

	payslip.OriginalFileName = fileName
	payslip.OriginalFileMime = mimeType
	payslip.OriginalFileSize = int64(len(fileBytes))
	payslip.OriginalFileContent = fileBytes
	payslip.ContentHash = contentHash

	if err := s.repo.Save(ctx, &payslip); err != nil {
		return nil, fmt.Errorf("failed to save payslip: %w", err)
	}

	s.logger.Info("Successfully imported payslip", "id", payslip.ID, "month", payslip.PeriodMonthNum, "name", payslip.EmployeeName)
	return &payslip, nil
}

func (s *PayslipService) Update(ctx context.Context, payslip *entity.Payslip) error {
	if len(payslip.OriginalFileContent) > 0 {
		hashFunc := sha256.New()
		hashFunc.Write(payslip.OriginalFileContent)
		payslip.ContentHash = hex.EncodeToString(hashFunc.Sum(nil))

		exists, err := s.repo.ExistsByHash(ctx, payslip.ContentHash)
		if err != nil {
			return fmt.Errorf("failed to check hash: %w", err)
		}
		if exists {
			return fmt.Errorf("%w: %s", ErrPayslipDuplicate, payslip.ContentHash)
		}
	}

	return s.repo.Update(ctx, payslip)
}

func (s *PayslipService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// JSONPayslipEntry represents a single record in the payslip bulk-import JSON manifest.
type JSONPayslipEntry struct {
	PeriodMonthNum   int            `json:"period_month_num"`
	PeriodYear       int            `json:"period_year"`
	EmployeeName     string         `json:"employee_name"`
	TaxClass         string         `json:"tax_class"`
	TaxID            string         `json:"tax_id"`
	GrossPay         float64        `json:"gross_pay"`
	NetPay           float64        `json:"net_pay"`
	PayoutAmount     float64        `json:"payout_amount"`
	CustomDeductions *float64       `json:"custom_deductions"`
	Bonuses          []entity.Bonus `json:"bonuses"`
	OriginalFileName string         `json:"original_file_name"`
}

func (s *PayslipService) ImportFromJSONFile(ctx context.Context, jsonFilePath string) (imported int, skipped int, errs []error, fatalErr error) {
	data, err := os.ReadFile(jsonFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil, nil
		}
		return 0, 0, nil, fmt.Errorf("payslip json import: read file: %w", err)
	}

	var entries []JSONPayslipEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return 0, 0, nil, fmt.Errorf("payslip json import: unmarshal: %w", err)
	}

	jsonDir := filepath.Dir(jsonFilePath)

	s.logger.Info("Payslip JSON import: starting", "file", jsonFilePath, "total_entries", len(entries))

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return imported, skipped, errs, ctx.Err()
		default:
		}

		if entry.OriginalFileName == "" {
			errs = append(errs, fmt.Errorf("skipping entry %d/%d: missing original_file_name", entry.PeriodMonthNum, entry.PeriodYear))
			continue
		}

		exists, checkErr := s.repo.ExistsByOriginalFileName(ctx, entry.OriginalFileName)
		if checkErr != nil {
			errs = append(errs, fmt.Errorf("check exists %s: %w", entry.OriginalFileName, checkErr))
			continue
		}
		if exists {
			s.logger.Debug("Payslip JSON import: skipping already-imported entry", "file", entry.OriginalFileName)
			skipped++
			continue
		}

		pdfPath := filepath.Join(jsonDir, entry.OriginalFileName)
		fileBytes, readErr := os.ReadFile(pdfPath)

		var contentHash string
		var fileMime string
		var fileSize int64

		if readErr == nil {
			h := sha256.Sum256(fileBytes)
			contentHash = hex.EncodeToString(h[:])
			fileMime = "application/pdf"
			fileSize = int64(len(fileBytes))
		} else {
			hashInput := fmt.Sprintf("json-import:%s:%d-%02d", entry.OriginalFileName, entry.PeriodYear, entry.PeriodMonthNum)
			h := sha256.Sum256([]byte(hashInput))
			contentHash = hex.EncodeToString(h[:])
		}

		hashExists, checkErr := s.repo.ExistsByHash(ctx, contentHash)
		if checkErr != nil {
			errs = append(errs, fmt.Errorf("check hash %s: %w", entry.OriginalFileName, checkErr))
			continue
		}
		if hashExists {
			skipped++
			continue
		}

		p := entity.Payslip{
			OriginalFileName:    entry.OriginalFileName,
			OriginalFileMime:    fileMime,
			OriginalFileSize:    fileSize,
			OriginalFileContent: fileBytes,
			ContentHash:         contentHash,
			PeriodMonthNum:      entry.PeriodMonthNum,
			PeriodYear:          entry.PeriodYear,
			EmployeeName:        entry.EmployeeName,
			TaxClass:            entry.TaxClass,
			TaxID:               entry.TaxID,
			GrossPay:            entry.GrossPay,
			NetPay:              entry.NetPay,
			PayoutAmount:        entry.PayoutAmount,
			Bonuses:             entry.Bonuses,
		}
		if entry.CustomDeductions != nil {
			p.CustomDeductions = *entry.CustomDeductions
		}

		if saveErr := s.repo.Save(ctx, &p); saveErr != nil {
			errs = append(errs, fmt.Errorf("save %s: %w", entry.OriginalFileName, saveErr))
			continue
		}

		s.logger.Info("Payslip JSON import: saved entry",
			"id", p.ID,
			"file", entry.OriginalFileName,
			"period", fmt.Sprintf("%d-%02d", entry.PeriodYear, entry.PeriodMonthNum),
			"file_stored", len(fileBytes) > 0,
		)
		imported++

		if len(fileBytes) > 0 {
			if removeErr := os.Remove(pdfPath); removeErr != nil && !os.IsNotExist(removeErr) {
				s.logger.Warn("Payslip JSON import: could not delete source PDF",
					"path", pdfPath, "error", removeErr)
			} else if removeErr == nil {
				s.logger.Info("Payslip JSON import: deleted source PDF", "path", pdfPath)
			}
		}
	}

	s.logger.Info("Payslip JSON import: finished",
		"imported", imported, "skipped", skipped, "errors", len(errs),
	)
	return imported, skipped, errs, nil
}
