package user

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDefaultService_Register(t *testing.T) {
	pepper := []byte("test-pepper")
	jwtKey := []byte("test-jwt-key-32-characters!!")

	t.Run("success", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		mockStore.CreateFunc = func(ctx context.Context, email, username, passwordHash, salt string, iterations int64) (*User, error) {
			return &User{
				ID:        uuid.New(),
				Email:     email,
				Username:  username,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}

		user, err := svc.Register(context.Background(), "test@example.com", "testuser", "password123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "testuser", user.Username)
		assert.NotEmpty(t, mockStore.CreateCalls())
	})

	t.Run("store error", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		mockStore.CreateFunc = func(ctx context.Context, email, username, passwordHash, salt string, iterations int64) (*User, error) {
			return nil, assert.AnError
		}

		user, err := svc.Register(context.Background(), "test@example.com", "testuser", "password123")

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestDefaultService_ValidatePassword(t *testing.T) {
	pepper := []byte("test-pepper")
	jwtKey := []byte("test-jwt-key-32-characters!!")
	testUserID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		// Generate a valid hash for "password123"
		hashResult, err := svc.hashPassword("password123")
		assert.NoError(t, err)

		mockStore.GetByEmailFunc = func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:             testUserID,
				Email:          email,
				Username:       "testuser",
				HashedPassword: []byte(hashResult.Hash),
				Salt:           []byte(hashResult.Salt),
				Iterations:     int64(hashResult.Iterations),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}, nil
		}

		user, err := svc.ValidatePassword(context.Background(), "test@example.com", "password123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUserID, user.ID)
	})

	t.Run("user not found", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		mockStore.GetByEmailFunc = func(ctx context.Context, email string) (*User, error) {
			return nil, assert.AnError
		}

		user, err := svc.ValidatePassword(context.Background(), "test@example.com", "password")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCreds, err)
		assert.Nil(t, user)
	})

	t.Run("invalid password", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		// Generate a valid hash for "password123"
		hashResult, err := svc.hashPassword("password123")
		assert.NoError(t, err)

		mockStore.GetByEmailFunc = func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:             testUserID,
				Email:          email,
				Username:       "testuser",
				HashedPassword: []byte(hashResult.Hash),
				Salt:           []byte(hashResult.Salt),
				Iterations:     int64(hashResult.Iterations),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}, nil
		}

		user, err := svc.ValidatePassword(context.Background(), "test@example.com", "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCreds, err)
		assert.Nil(t, user)
	})
}

func TestDefaultService_GenerateToken(t *testing.T) {
	pepper := []byte("test-pepper")
	jwtKey := []byte("test-jwt-key-32-characters!!")
	testUser := &User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		token, err := svc.GenerateToken(testUser)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestDefaultService_hashPassword(t *testing.T) {
	pepper := []byte("test-pepper")
	jwtKey := []byte("test-jwt-key-32-characters!!")

	t.Run("generates valid hash", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		result, err := svc.hashPassword("password123")

		assert.NoError(t, err)
		assert.NotEmpty(t, result.Hash)
		assert.NotEmpty(t, result.Salt)
		assert.Equal(t, uint32(DefaultIterations), result.Iterations)
		assert.Equal(t, uint32(DefaultMemory), result.Memory)
		assert.Equal(t, uint8(DefaultParallelism), result.Threads)
	})

	t.Run("different salts produce different hashes", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		result1, _ := svc.hashPassword("password123")
		result2, _ := svc.hashPassword("password123")

		assert.NotEqual(t, result1.Salt, result2.Salt)
		assert.NotEqual(t, result1.Hash, result2.Hash)
	})
}

func TestDefaultService_verifyPassword(t *testing.T) {
	pepper := []byte("test-pepper")
	jwtKey := []byte("test-jwt-key-32-characters!!")

	t.Run("correct password returns true", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		result, _ := svc.hashPassword("password123")
		valid := svc.verifyPassword("password123", result.Salt, result.Hash, result.Iterations)

		assert.True(t, valid)
	})

	t.Run("incorrect password returns false", func(t *testing.T) {
		mockStore := &PgxStoreMock{}
		svc := NewDefaultService(mockStore, pepper, 0, 0, jwtKey)

		result, _ := svc.hashPassword("password123")
		valid := svc.verifyPassword("wrongpassword", result.Salt, result.Hash, result.Iterations)

		assert.False(t, valid)
	})
}
