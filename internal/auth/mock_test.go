package auth

import (
	"context"
	"time"
)

type mockRepository struct {
	createFn     func(ctx context.Context, user *User) error
	findByEmail  func(ctx context.Context, email string) (*User, error)
	findByID     func(ctx context.Context, id string) (*User, error)
	listAllFn    func(ctx context.Context, role string, limit int, offset int) ([]User, int, error)
	updateRoleFn func(ctx context.Context, id string, role string) error
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

func (m *mockRepository) ListAll(ctx context.Context, role string, limit int, offset int) ([]User, int, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx, role, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockRepository) UpdateRole(ctx context.Context, id string, role string) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, id, role)
	}
	return nil
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
