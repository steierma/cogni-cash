package entity

// Setting represents a key-value configuration pair stored in the database.
type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
