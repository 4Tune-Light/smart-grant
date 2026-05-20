package audit

import "time"

type Audit struct {
	ID         string
	EntityType string
	EntityID   string
	Action     string
	ActorID    string
	OldValues  string
	NewValues  string
	CreatedAt  time.Time
}
