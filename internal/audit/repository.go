package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rizky/smart-grant/pkg/cursor"
)

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

type Repository interface {
	Insert(ctx context.Context, entry *Audit) error
	List(ctx context.Context, filter AuditFilter) ([]Audit, *cursor.Cursor, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Insert(ctx context.Context, entry *Audit) error {
	query := `
		INSERT INTO audit_logs (entity_type, entity_id, action, actor_id, old_values, new_values)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		entry.EntityType, entry.EntityID, entry.Action, entry.ActorID,
		nullIfEmpty(entry.OldValues), nullIfEmpty(entry.NewValues),
	).Scan(&entry.ID, &entry.CreatedAt)
}

func (r *repository) List(ctx context.Context, filter AuditFilter) ([]Audit, *cursor.Cursor, error) {
	query := "SELECT id, entity_type, entity_id, action, actor_id, COALESCE(old_values, '{}'::jsonb), COALESCE(new_values, '{}'::jsonb), created_at FROM audit_logs"

	args := []interface{}{}
	where := ""
	argIdx := 1

	if filter.EntityType != "" {
		where += fmt.Sprintf(" WHERE entity_type = $%d", argIdx)
		args = append(args, filter.EntityType)
		argIdx++
	}
	if filter.EntityID != "" {
		prefix := " AND"
		if where == "" {
			prefix = " WHERE"
		}
		where += fmt.Sprintf("%s entity_id = $%d", prefix, argIdx)
		args = append(args, filter.EntityID)
		argIdx++
	}
	if filter.ActorID != "" {
		prefix := " AND"
		if where == "" {
			prefix = " WHERE"
		}
		where += fmt.Sprintf("%s actor_id = $%d", prefix, argIdx)
		args = append(args, filter.ActorID)
		argIdx++
	}
	if filter.Action != "" {
		prefix := " AND"
		if where == "" {
			prefix = " WHERE"
		}
		where += fmt.Sprintf("%s action = $%d", prefix, argIdx)
		args = append(args, filter.Action)
		argIdx++
	}

	query += where

	if filter.Cursor != nil {
		prefix := " AND"
		if where == "" {
			prefix = " WHERE"
		}
		query += fmt.Sprintf("%s (created_at, id) < ($%d::timestamptz, $%d::uuid)", prefix, argIdx, argIdx+1)
		args = append(args, filter.Cursor.LastCreatedAt, filter.Cursor.LastID)
		argIdx += 2
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", argIdx)
	args = append(args, filter.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var entries []Audit
	for rows.Next() {
		var e Audit
		if err := rows.Scan(&e.ID, &e.EntityType, &e.EntityID, &e.Action, &e.ActorID, &e.OldValues, &e.NewValues, &e.CreatedAt); err != nil {
			return nil, nil, err
		}
		entries = append(entries, e)
	}

	var nextCursor *cursor.Cursor
	hasMore := len(entries) > filter.Limit
	if hasMore {
		entries = entries[:filter.Limit]
		last := entries[len(entries)-1]
		nextCursor = &cursor.Cursor{LastID: last.ID, LastCreatedAt: last.CreatedAt}
	}

	return entries, nextCursor, nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" || s == "{}" || s == `""` {
		return nil
	}
	return s
}
