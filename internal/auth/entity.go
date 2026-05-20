package auth

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	Role         string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) CanLogin() error {
	if !u.IsActive {
		return ErrUserInactive
	}
	return nil
}
