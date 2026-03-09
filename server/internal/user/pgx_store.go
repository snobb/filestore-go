package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	HashedPassword []byte    `json:"hashed_password"`
	Salt           []byte    `json:"salt"`
	Iterations     int64     `json:"iterations"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

//go:generate moq -rm -fmt goimports -out pgx_store_mock.go . PgxStore

type PgxStore interface {
	Create(ctx context.Context, email, username, passwordHash, salt string, iterations int64) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}
