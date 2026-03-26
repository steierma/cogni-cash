package entity

// ReconciliationPairSuggestion represents a proposed 1:1 match between two
// transactions from different accounts (e.g. Giro settlement ↔ Visa credit).
type ReconciliationPairSuggestion struct {
	SourceTransaction Transaction `json:"source_transaction"`
	TargetTransaction Transaction `json:"target_transaction"`
	MatchScore        float64     `json:"match_score"`
}

