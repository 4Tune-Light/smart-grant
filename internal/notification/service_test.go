package notification

import (
	"context"
	"testing"

	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	insertFn     func(ctx context.Context, n *Notification) error
	findByUserFn func(ctx context.Context, userID string, limit int, c *cursor.Cursor) ([]Notification, *cursor.Cursor, error)
	markReadFn   func(ctx context.Context, id string, userID string) error
}

func (m *mockRepository) Insert(ctx context.Context, n *Notification) error { return m.insertFn(ctx, n) }
func (m *mockRepository) FindByUserID(ctx context.Context, userID string, limit int, c *cursor.Cursor) ([]Notification, *cursor.Cursor, error) { return m.findByUserFn(ctx, userID, limit, c) }
func (m *mockRepository) MarkRead(ctx context.Context, id string, userID string) error { return m.markReadFn(ctx, id, userID) }

func authCtx(userID, role string) context.Context {
	ctx := context.WithValue(context.Background(), middleware.AuthUserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.AuthRoleKey, role)
	return ctx
}

func TestSend_Success(t *testing.T) {
	repo := &mockRepository{
		insertFn: func(ctx context.Context, n *Notification) error { return nil },
	}
	svc := NewService(repo, nil)

	err := svc.Send(context.Background(), "user-1", "test", "Hello", "Test body")
	assert.NoError(t, err)
}

func TestList_ReturnsByUser(t *testing.T) {
	repo := &mockRepository{
		findByUserFn: func(ctx context.Context, userID string, limit int, c *cursor.Cursor) ([]Notification, *cursor.Cursor, error) {
			assert.Equal(t, "user-1", userID)
			return []Notification{
				{UserID: "user-1", Type: "test", Title: "Hello", Body: "Body"},
			}, nil, nil
		},
	}
	svc := NewService(repo, nil)

	ctx := authCtx("user-1", "applicant")
	resp, _, err := svc.List(ctx, 10, nil)

	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "Hello", resp[0].Title)
}

func TestMarkRead_Success(t *testing.T) {
	repo := &mockRepository{
		markReadFn: func(ctx context.Context, id string, userID string) error {
			assert.Equal(t, "notif-1", id)
			return nil
		},
	}
	svc := NewService(repo, nil)

	ctx := authCtx("user-1", "applicant")
	err := svc.MarkRead(ctx, "notif-1")
	assert.NoError(t, err)
}
