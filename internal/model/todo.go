package model

// Todo represents a single todo item. Fields are stubs for now;
// full implementation comes in issue #2.
type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}
