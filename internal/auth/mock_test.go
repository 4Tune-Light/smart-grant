package auth

import (
	"context"
	"time"
)

type mockRepository struct {
	createFn    func(ctx context.Context, user *User) error
	findByEmail func(ctx context.Context, email string) (*User, error)
	findByID    func(ctx context.Context, id string) (*User, error)
}

func (m *mockRepository) Create(ctx context.Context, user *User) error {
	return m.createFn(ctx, user)
}

func (m *mockRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	return m.findByEmail(ctx, email)
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*User, error) {
	return m.findByID(ctx, id)
}

func testUser(id, email, role string) *User {
	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: "$argon2id$v=19$m=65536,t=1,p=4$salt$hash",
		Name:         "Test User",
		Role:         role,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
