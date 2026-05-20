package notification

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/rizky/smart-grant/pkg/database"
)

type Repository interface {
	Insert(ctx context.Context, n *Notification) error
	FindByUserID(ctx context.Context, userID string, limit int, c *cursor.Cursor) ([]Notification, *cursor.Cursor, error)
	MarkRead(ctx context.Context, id string, userID string) error
}

type repository struct {
	q *database.Querier
}

func NewRepository(q *database.Querier) Repository {
	return &repository{q: q}
}

func (r *repository) Insert(ctx context.Context, n *Notification) error {
	query := `
		INSERT INTO notifications (user_id, type, title, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return r.q.QueryRow(ctx, query, n.UserID, n.Type, n.Title, n.Body).Scan(&n.ID, &n.CreatedAt)
}

func (r *repository) FindByUserID(ctx context.Context, userID string, limit int, c *cursor.Cursor) ([]Notification, *cursor.Cursor, error) {
	query := `
		SELECT id, user_id, type, title, body, is_read, created_at
		FROM notifications WHERE user_id = $1`

	args := []interface{}{userID}
	argIdx := 2

	if c != nil {
		query += fmt.Sprintf(` AND (created_at, id) < ($%d::timestamptz, $%d::uuid)`, argIdx, argIdx+1)
		args = append(args, c.LastCreatedAt, c.LastID)
		argIdx += 2
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d`, argIdx)
	args = append(args, limit+1)

	rows, err := r.q.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, nil, err
		}
		notifications = append(notifications, n)
	}

	var nextCursor *cursor.Cursor
	hasMore := len(notifications) > limit
	if hasMore {
		notifications = notifications[:limit]
		last := notifications[len(notifications)-1]
		nextCursor = &cursor.Cursor{LastID: last.ID, LastCreatedAt: last.CreatedAt}
	}

	return notifications, nextCursor, nil
}

func (r *repository) MarkRead(ctx context.Context, id string, userID string) error {
	query := `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`
	result, err := r.q.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
