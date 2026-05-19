package audit

import (
	"context"
	"encoding/json"

	"github.com/rizky/smart-grant/pkg/cursor"
)

type Service interface {
	Log(ctx context.Context, entry LogEntry) error
	List(ctx context.Context, filter AuditFilter) ([]AuditResponse, *cursor.Cursor, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Log(ctx context.Context, entry LogEntry) error {
	oldStr := ""
	if entry.OldValues != "" {
		oldStr = entry.OldValues
	}
	newStr := ""
	if entry.NewValues != "" {
		newStr = entry.NewValues
	}

	a := &Audit{
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		Action:     entry.Action,
		ActorID:    entry.ActorID,
		OldValues:  oldStr,
		NewValues:  newStr,
	}

	return s.repo.Insert(ctx, a)
}

func (s *service) List(ctx context.Context, filter AuditFilter) ([]AuditResponse, *cursor.Cursor, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	entries, nextCursor, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]AuditResponse, len(entries))
	for i, e := range entries {
		responses[i] = AuditResponse{
			ID:         e.ID,
			EntityType: e.EntityType,
			EntityID:   e.EntityID,
			Action:     e.Action,
			ActorID:    e.ActorID,
			OldValues:  formatJSON(e.OldValues),
			NewValues:  formatJSON(e.NewValues),
			CreatedAt:  e.CreatedAt,
		}
	}

	return responses, nextCursor, nil
}

func formatJSON(s string) string {
	if s == "" {
		return ""
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	return s
}
