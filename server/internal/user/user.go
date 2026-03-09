// Package user defines interface to user information.
// XXX: I leave the APIs but leave unimplemented because auth is out of scope.
package user

import (
	"context"
	"time"

	"server/internal/pglib"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	HashedPassword string    `json:"hashed_password"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserStore struct {
	db pglib.Pool
}

func New(db pglib.Pool) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, email, username, hashedPassword string) (*User, error) {
	panic("not implemented")
}

func (s *UserStore) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	panic("not implemented")
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	panic("not implemented")
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	panic("not implemented")
}

func (s *UserStore) List(ctx context.Context) ([]User, error) {
	panic("not implemented")
}
