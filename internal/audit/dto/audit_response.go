package dto

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
