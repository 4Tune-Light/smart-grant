package audit

import (
	"context"
	"testing"

	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	insertFn func(ctx context.Context, entry *Audit) error
	listFn   func(ctx context.Context, filter AuditFilter) ([]Audit, *cursor.Cursor, error)
}

func (m *mockRepository) Insert(ctx context.Context, entry *Audit) error {
	return m.insertFn(ctx, entry)
}

func (m *mockRepository) List(ctx context.Context, filter AuditFilter) ([]Audit, *cursor.Cursor, error) {
	return m.listFn(ctx, filter)
}

func TestLog_Success(t *testing.T) {
	repo := &mockRepository{
		insertFn: func(ctx context.Context, entry *Audit) error { return nil },
	}
	svc := NewService(repo)

	err := svc.Log(context.Background(), LogEntry{
		EntityType: "proposal", EntityID: "p1",
		Action: "create", ActorID: "user-1",
	})

	assert.NoError(t, err)
}

func TestList_ReturnsEntries(t *testing.T) {
	repo := &mockRepository{
		listFn: func(ctx context.Context, filter AuditFilter) ([]Audit, *cursor.Cursor, error) {
			return []Audit{
				{EntityType: "proposal", EntityID: "p1", Action: "create", ActorID: "u1"},
			}, nil, nil
		},
	}
	svc := NewService(repo)

	resp, nextCursor, err := svc.List(context.Background(), AuditFilter{Limit: 20})

	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "p1", resp[0].EntityID)
	assert.Nil(t, nextCursor)
}

func TestList_FilterByEntityType(t *testing.T) {
	repo := &mockRepository{
		listFn: func(ctx context.Context, filter AuditFilter) ([]Audit, *cursor.Cursor, error) {
			assert.Equal(t, "proposal", filter.EntityType)
			return nil, nil, nil
		},
	}
	svc := NewService(repo)

	_, _, err := svc.List(context.Background(), AuditFilter{EntityType: "proposal", Limit: 20})
	assert.NoError(t, err)
}
