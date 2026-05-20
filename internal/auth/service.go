package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	authdto "github.com/rizky/smart-grant/internal/auth/dto"
	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
)

type TokenConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type Service interface {
	Register(ctx context.Context, req authdto.RegisterRequest) (*authdto.AuthResponse, error)
	Login(ctx context.Context, req authdto.LoginRequest) (*authdto.AuthResponse, error)
	RefreshToken(ctx context.Context, req authdto.RefreshRequest) (*authdto.AuthResponse, error)
	ListUsers(ctx context.Context, role string, limit int, page int) ([]authdto.UserInfo, int, error)
	UpdateRole(ctx context.Context, targetID string, newRole string) error
}

type service struct {
	repo  Repository
	token TokenConfig
}

func NewService(repo Repository, token TokenConfig) Service {
	return &service{repo: repo, token: token}
}

func (s *service) Register(ctx context.Context, req authdto.RegisterRequest) (*authdto.AuthResponse, error) {
	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         req.Role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateTokenResponse(user)
}

func (s *service) Login(ctx context.Context, req authdto.LoginRequest) (*authdto.AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	ok, err := verifyPassword(req.Password, user.PasswordHash)
	if err != nil || !ok {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenResponse(user)
}

func (s *service) RefreshToken(ctx context.Context, req authdto.RefreshRequest) (*authdto.AuthResponse, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.token.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return s.generateTokenResponse(user)
}

func (s *service) generateTokenResponse(user *User) (*authdto.AuthResponse, error) {
	now := time.Now()

	accessClaims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.token.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "smart-grant",
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.token.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := &Claims{
		UserID:    user.ID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.token.RefreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "smart-grant",
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.token.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &authdto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.token.AccessTTL.Seconds()),
		User: authdto.UserInfo{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
	}, nil
}

func (s *service) ListUsers(ctx context.Context, role string, limit int, page int) ([]authdto.UserInfo, int, error) {
	offset := (page - 1) * limit
	users, total, err := s.repo.ListAll(ctx, role, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	info := make([]authdto.UserInfo, len(users))
	for i, u := range users {
		info[i] = authdto.UserInfo{ID: u.ID, Email: u.Email, Name: u.Name, Role: u.Role}
	}
	return info, total, nil
}

func (s *service) UpdateRole(ctx context.Context, targetID string, newRole string) error {
	actorID, _ := ctx.Value("auth_user_id").(string)
	if actorID == targetID {
		return fmt.Errorf("cannot change your own role")
	}
	return s.repo.UpdateRole(ctx, targetID, newRole)
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads, b64Salt, b64Hash), nil
}

func verifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("invalid hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, errors.New("invalid hash version")
	}

	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, errors.New("invalid hash params")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, errors.New("invalid salt encoding")
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, errors.New("invalid hash encoding")
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1, nil
}

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}
