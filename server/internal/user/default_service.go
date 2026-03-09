package user

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrInvalidCreds = errors.New("invalid credentials")
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

const (
	TokenExpiry = 24 * time.Hour
)

func init() {
}

// Parameters for Argon2id (2026 Recommended Defaults)
const (
	DefaultMemory      = 64 * 1024 // 64MB
	DefaultIterations  = 3
	DefaultParallelism = 2
	KeyLen             = 32
	SaltLen            = 16
)

type PasswordResult struct {
	Hash       string
	Salt       string
	Iterations uint32
	Memory     uint32
	Threads    uint8
}

type DefaultService struct {
	pgxStore PgxStore
	pepper   []byte
	memory   uint32
	threads  uint8
	jwtKey   []byte
}

func NewDefaultService(pgxStore PgxStore, pepper []byte, memory uint32, threads uint8, jwtKey []byte) *DefaultService {
	if memory == 0 {
		memory = DefaultMemory
	}

	if threads == 0 {
		threads = DefaultParallelism
	}

	if jwtKey == nil {
		jwtKey = make([]byte, 32)
		rand.Read(jwtKey)
	}

	return &DefaultService{
		pgxStore: pgxStore,
		pepper:   pepper,
		memory:   memory,
		threads:  threads,
		jwtKey:   jwtKey,
	}
}

func (s *DefaultService) hashPassword(password string) (PasswordResult, error) {
	salt := make([]byte, SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return PasswordResult{}, err
	}

	// We append the pepper to the password
	hash := argon2.IDKey(
		[]byte(password+string(s.pepper)),
		salt,
		DefaultIterations,
		DefaultMemory,
		DefaultParallelism,
		KeyLen,
	)

	return PasswordResult{
		Hash:       hex.EncodeToString(hash),
		Salt:       hex.EncodeToString(salt),
		Iterations: DefaultIterations,
		Memory:     DefaultMemory,
		Threads:    DefaultParallelism,
	}, nil
}

func (s *DefaultService) verifyPassword(password, salt, hash string, iterations uint32) bool {
	// Decode stored hex values
	saltBytes, _ := hex.DecodeString(salt)
	originalHash, _ := hex.DecodeString(hash)

	// Derive hash from attempt using stored parameters
	candidateHash := argon2.IDKey(
		[]byte(password+string(s.pepper)),
		saltBytes,
		iterations,
		s.memory,
		s.threads,
		uint32(len(originalHash)),
	)

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(originalHash, candidateHash) == 1
}

func (s *DefaultService) Register(ctx context.Context, email, username, password string) (*User, error) {
	passwordResult, err := s.hashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.pgxStore.Create(
		ctx,
		email,
		username,
		passwordResult.Hash,
		passwordResult.Salt,
		int64(passwordResult.Iterations),
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *DefaultService) ValidatePassword(ctx context.Context, email, password string) (*User, error) {
	user, err := s.pgxStore.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCreds
	}

	saltHex := string(user.Salt)
	hashHex := string(user.HashedPassword)

	if !s.verifyPassword(password, saltHex, hashHex, uint32(user.Iterations)) {
		return nil, ErrInvalidCreds
	}

	return user, nil
}

func (s *DefaultService) GenerateToken(user *User) (string, error) {
	claims := JWTClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtKey)
}
