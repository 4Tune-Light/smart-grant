package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func tokenConfig() TokenConfig {
	return TokenConfig{
		Secret:     "test-secret-that-is-at-least-32-bytes-long!!",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 72 * time.Hour,
	}
}

func authCtx(userID, role string) context.Context {
	ctx := context.WithValue(context.Background(), middleware.AuthUserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.AuthRoleKey, role)
	return ctx
}

func TestRegister_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, user *User) error {
			user.ID = "new-uuid"
			user.CreatedAt = time.Now()
			user.UpdatedAt = time.Now()
			return nil
		},
	}
	svc := NewService(repo, tokenConfig())

	resp, err := svc.Register(context.Background(), RegisterRequest{
		Email: "test@example.com", Password: "password123",
		Name: "Test", Role: "applicant",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "test@example.com", resp.User.Email)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, user *User) error {
			return ErrEmailAlreadyExists
		},
	}
	svc := NewService(repo, tokenConfig())

	_, err := svc.Register(context.Background(), RegisterRequest{
		Email: "dup@example.com", Password: "password123",
		Name: "Dup", Role: "applicant",
	})

	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
}

func TestLogin_Success(t *testing.T) {
	repo := &mockRepository{
		findByEmail: func(ctx context.Context, email string) (*User, error) {
			u := testUser("u1", email, "applicant")
			hash, _ := hashPassword("correct-password")
			u.PasswordHash = hash
			return u, nil
		},
	}
	svc := NewService(repo, tokenConfig())

	resp, err := svc.Login(context.Background(), LoginRequest{
		Email: "test@example.com", Password: "correct-password",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &mockRepository{
		findByEmail: func(ctx context.Context, email string) (*User, error) {
			u := testUser("u1", email, "applicant")
			hash, _ := hashPassword("correct-password")
			u.PasswordHash = hash
			return u, nil
		},
	}
	svc := NewService(repo, tokenConfig())

	_, err := svc.Login(context.Background(), LoginRequest{
		Email: "test@example.com", Password: "wrong-password",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_InactiveUser(t *testing.T) {
	repo := &mockRepository{
		findByEmail: func(ctx context.Context, email string) (*User, error) {
			u := testUser("u1", email, "applicant")
			u.IsActive = false
			return u, nil
		},
	}
	svc := NewService(repo, tokenConfig())

	_, err := svc.Login(context.Background(), LoginRequest{
		Email: "test@example.com", Password: "any-password",
	})

	assert.ErrorIs(t, err, ErrUserInactive)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockRepository{
		findByEmail: func(ctx context.Context, email string) (*User, error) {
			return nil, ErrUserNotFound
		},
	}
	svc := NewService(repo, tokenConfig())

	_, err := svc.Login(context.Background(), LoginRequest{
		Email: "nonexistent@example.com", Password: "any-password",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestRefreshToken_Valid(t *testing.T) {
	repo := &mockRepository{
		findByID: func(ctx context.Context, id string) (*User, error) {
			return testUser(id, "test@example.com", "applicant"), nil
		},
	}
	svc := NewService(repo, tokenConfig())

	claims := &Claims{
		UserID: "u1", TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Issuer:    "smart-grant",
		},
	}
	refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(tokenConfig().Secret))

	resp, err := svc.RefreshToken(context.Background(), RefreshRequest{RefreshToken: refreshToken})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestRefreshToken_Expired(t *testing.T) {
	svc := NewService(&mockRepository{}, tokenConfig())

	claims := &Claims{
		UserID: "u1", TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			Issuer:    "smart-grant",
		},
	}
	refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(tokenConfig().Secret))

	_, err := svc.RefreshToken(context.Background(), RefreshRequest{RefreshToken: refreshToken})

	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestRefreshToken_WrongType(t *testing.T) {
	svc := NewService(&mockRepository{}, tokenConfig())

	claims := &Claims{
		UserID: "u1", TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Issuer:    "smart-grant",
		},
	}
	refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(tokenConfig().Secret))

	_, err := svc.RefreshToken(context.Background(), RefreshRequest{RefreshToken: refreshToken})

	assert.ErrorIs(t, err, ErrInvalidToken)
}
