package audit

import "github.com/rizky/smart-grant/pkg/cursor"

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
	Limit      int
	Cursor     *cursor.Cursor
}
