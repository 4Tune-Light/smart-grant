package notification

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	Insert(ctx context.Context, n *Notification) error
	FindByUserID(ctx context.Context, userID string, limit int, offset int) ([]Notification, int, error)
	MarkRead(ctx context.Context, id string, userID string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Insert(ctx context.Context, n *Notification) error {
	query := `
		INSERT INTO notifications (user_id, type, title, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, n.UserID, n.Type, n.Title, n.Body).Scan(&n.ID, &n.CreatedAt)
}

func (r *repository) FindByUserID(ctx context.Context, userID string, limit int, offset int) ([]Notification, int, error) {
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, type, title, body, is_read, created_at
		FROM notifications WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, n)
	}

	return notifications, total, nil
}

func (r *repository) MarkRead(ctx context.Context, id string, userID string) error {
	query := `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`
	result, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
