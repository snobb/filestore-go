package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"server/internal/pglib"
)

type DefaultPgxStore struct {
	db pglib.Pool
}

func NewDefaultPgxStore(db pglib.Pool) *DefaultPgxStore {
	return &DefaultPgxStore{db: db}
}

func (s *DefaultPgxStore) Create(
	ctx context.Context,
	email, username, passwordHash, salt string,
	iterations int64,
) (*User, error) {
	var user User
	var pgUserID pgtype.UUID

	err := s.db.QueryRow(ctx, `
		INSERT INTO users (email, username, hashed_password, salt, iterations)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email)
		DO UPDATE SET
			username=EXCLUDED.username,
			hashed_password=EXCLUDED.hashed_password,
			salt=EXCLUDED.salt,
			iterations=EXCLUDED.iterations,
			updated_at = NOW()
		RETURNING id, email, username, hashed_password, salt, iterations, created_at, updated_at
	`, email, username, passwordHash, salt, iterations).Scan(
		&pgUserID, &user.Email, &user.Username, &user.HashedPassword,
		&user.Salt, &user.Iterations, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	userID, err := pglib.ParsePgUUID(pgUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID %v > %w", pgUserID, err)
	}

	user.ID = userID
	return &user, nil
}

func (s *DefaultPgxStore) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	var pgUserID pgtype.UUID
	err := s.db.QueryRow(ctx, `
		SELECT id, email, username, hashed_password, salt, iterations, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&pgUserID, &user.Email, &user.Username, &user.HashedPassword,
		&user.Salt, &user.Iterations, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	userID, err := pglib.ParsePgUUID(pgUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID > %w", err)
	}

	user.ID = userID
	return &user, nil
}

func (s *DefaultPgxStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	var pgUserID pgtype.UUID
	err := s.db.QueryRow(ctx, `
		SELECT id, email, username, hashed_password, salt, iterations, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&pgUserID, &user.Email, &user.Username, &user.HashedPassword,
		&user.Salt, &user.Iterations, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	userID, err := pglib.ParsePgUUID(pgUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID > %w", err)
	}

	user.ID = userID
	return &user, nil
}
