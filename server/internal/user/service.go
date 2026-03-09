package user

import (
	"context"
)

//go:generate moq -rm -fmt goimports -out service_mock.go . Service

type Service interface {
	Register(ctx context.Context, email, username, password string) (*User, error)
	ValidatePassword(ctx context.Context, email, password string) (*User, error)
	GenerateToken(user *User) (string, error)
}
