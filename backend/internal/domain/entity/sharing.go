package entity

import (
	"github.com/google/uuid"
)

// SharingDashboard provides a consolidated view of all shared items and balances.
type SharingDashboard struct {
	SharedCategories []SharedCategorySummary `json:"shared_categories"`
	SharedInvoices   []Invoice               `json:"shared_invoices"`
	Balances         []CategoryBalance       `json:"balances"`
}

// SharedCategorySummary wraps a category with its sharing status.
type SharedCategorySummary struct {
	Category
	Permissions string `json:"permissions"` // "view" or "edit"
}

// CategoryBalance represents the "Who Paid What" logic for a shared category.
type CategoryBalance struct {
	CategoryID    uuid.UUID      `json:"category_id"`
	CategoryName  string         `json:"category_name"`
	TotalSpent    float64        `json:"total_spent"`
	UserBreakdown []UserSpending `json:"user_breakdown"`
}

// UserSpending tracks how much a specific user spent in a shared category.
type UserSpending struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Amount   float64   `json:"amount"`
}
