package audit

import "time"

type AuditResponse struct {
	ID         string    `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Action     string    `json:"action"`
	ActorID    string    `json:"actor_id"`
	OldValues  string    `json:"old_values,omitempty"`
	NewValues  string    `json:"new_values,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type LogEntry struct {
	EntityType string
	EntityID   string
	Action     string
	ActorID    string
	OldValues  string
	NewValues  string
}

type AuditFilter struct {
	EntityType string
	EntityID   string
	ActorID    string
	Action     string
	Page       int
	Limit      int
}
