package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

func (s *DiscoveryService) PreviewCancellation(ctx context.Context, userID, subID uuid.UUID, language string) (port.CancellationLetterResult, error) {
	// 1. Fetch Subscription
	sub, err := s.subRepo.GetByID(ctx, subID, userID)
	if err != nil {
		return port.CancellationLetterResult{}, err
	}

	// 2. Fetch User
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return port.CancellationLetterResult{}, err
	}

	// 3. Generate Draft
	endDate := ""
	if sub.ContractEndDate != nil {
		endDate = sub.ContractEndDate.Format("2006-01-02")
	}

	if language == "" {
		language = "DE" // Default
	}

	custNum := ""
	if sub.CustomerNumber != nil {
		custNum = *sub.CustomerNumber
	}
	noticeDays := 30
	if sub.NoticePeriodDays != nil {
		noticeDays = *sub.NoticePeriodDays
	}

	req := port.CancellationLetterRequest{
		UserFullName:     user.FullName,
		UserEmail:        user.Email,
		MerchantName:     sub.MerchantName,
		CustomerNumber:   custNum,
		ContractEndDate:  endDate,
		NoticePeriodDays: noticeDays,
		Language:         language,
	}

	return s.letterGen.GenerateCancellationLetter(ctx, userID, req)
}

func (s *DiscoveryService) CancelSubscription(ctx context.Context, userID, subID uuid.UUID, subject, body string) error {
	// 1. Fetch Subscription
	sub, err := s.subRepo.GetByID(ctx, subID, userID)
	if err != nil {
		return err
	}

	if sub.ContactEmail == nil || *sub.ContactEmail == "" {
		return errors.New("merchant contact email is missing")
	}

	// 2. Send Email
	err = s.email.Send(ctx, userID, *sub.ContactEmail, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send cancellation email: %w", err)
	}

	// 3. Update Status
	sub.Status = entity.SubscriptionStatusCancellationPending
	sub.UpdatedAt = time.Now()
	_, err = s.subRepo.Update(ctx, sub)
	if err != nil {
		s.Logger.Error("Failed to update subscription status after cancellation", "error", err, "sub_id", subID)
	}

	// 4. Log Event
	event := entity.SubscriptionEvent{
		SubscriptionID: subID,
		UserID:         userID,
		EventType:      "cancellation_sent",
		Title:          "Cancellation Email Sent",
		Content:        fmt.Sprintf("To: %s\nSubject: %s\n\n%s", *sub.ContactEmail, subject, body),
	}
	_ = s.subRepo.LogEvent(ctx, event)

	return nil
}
