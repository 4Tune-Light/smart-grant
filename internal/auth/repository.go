package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rizky/smart-grant/pkg/database"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	ListAll(ctx context.Context, role string, limit int, offset int) ([]User, int, error)
	UpdateRole(ctx context.Context, id string, role string) error
}

type repository struct {
	q *database.Querier
}

func NewRepository(q *database.Querier) Repository {
	return &repository{q: q}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (email, password_hash, name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.q.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.Name, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if pgErrCode(err) == "23505" {
			return ErrEmailAlreadyExists
		}
		return err
	}

	user.IsActive = true
	return nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, role, is_active, created_at, updated_at
		FROM users WHERE email = $1`

	user := &User{}
	err := r.q.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.Name, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, role, is_active, created_at, updated_at
		FROM users WHERE id = $1`

	user := &User{}
	err := r.q.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.Name, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *repository) ListAll(ctx context.Context, role string, limit int, offset int) ([]User, int, error) {
	countQuery := "SELECT COUNT(*) FROM users"
	query := "SELECT id, email, name, role, is_active, created_at, updated_at FROM users"
	args := []interface{}{}

	if role != "" {
		countQuery += " WHERE role = $1"
		query += " WHERE role = $1"
		args = append(args, role)
	}

	var total int
	if len(args) > 0 {
		if err := r.q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.q.QueryRow(ctx, countQuery).Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	if role == "" {
		query = "SELECT id, email, name, role, is_active, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2"
		args = append(args, limit, offset)
	} else {
		args = append(args, limit, offset)
	}

	rows, err := r.q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (r *repository) UpdateRole(ctx context.Context, id string, role string) error {
	query := `UPDATE users SET role = $1, updated_at = now() WHERE id = $2`
	result, err := r.q.Exec(ctx, query, role, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func pgErrCode(err error) string {
	type sqlErr interface {
		SQLState() string
	}
	var se sqlErr
	if errors.As(err, &se) {
		return se.SQLState()
	}
	return ""
}
