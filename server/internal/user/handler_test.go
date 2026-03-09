package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name         string
		body         string
		setupService func(*ServiceMock)
		wantStatus   int
		wantErr      string
	}{
		{
			name: "success",
			body: `{"email": "test@example.com", "username": "testuser", "password": "password123"}`,
			setupService: func(m *ServiceMock) {
				m.RegisterFunc = func(ctx context.Context, email, username, password string) (*User, error) {
					return &User{
						ID:        testUserID,
						Email:     email,
						Username:  username,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				}
				m.GenerateTokenFunc = func(user *User) (string, error) {
					return "test-token", nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:         "invalid json",
			body:         `invalid`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "invalid request",
		},
		{
			name:         "missing email",
			body:         `{"username": "testuser", "password": "password123"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "email, username and password are required",
		},
		{
			name:         "missing username",
			body:         `{"email": "test@example.com", "password": "password123"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "email, username and password are required",
		},
		{
			name:         "missing password",
			body:         `{"email": "test@example.com", "username": "testuser"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "email, username and password are required",
		},
		{
			name: "service error",
			body: `{"email": "test@example.com", "username": "testuser", "password": "password123"}`,
			setupService: func(m *ServiceMock) {
				m.RegisterFunc = func(ctx context.Context, email, username, password string) (*User, error) {
					return nil, assert.AnError
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "unable to register user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &ServiceMock{}
			tt.setupService(mockSvc)

			h := NewHandler(mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.Register(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErr != "" {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], tt.wantErr)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	testUserID := uuid.New()
	testEmail := "test@example.com"

	tests := []struct {
		name         string
		body         string
		setupService func(*ServiceMock)
		wantStatus   int
		wantErr      string
		checkToken   bool
	}{
		{
			name: "success",
			body: `{"email": "test@example.com", "password": "password123"}`,
			setupService: func(m *ServiceMock) {
				m.ValidatePasswordFunc = func(ctx context.Context, email, password string) (*User, error) {
					return &User{
						ID:        testUserID,
						Email:     email,
						Username:  "testuser",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				}
				m.GenerateTokenFunc = func(user *User) (string, error) {
					return "test-token", nil
				}
			},
			wantStatus: http.StatusOK,
			checkToken: true,
		},
		{
			name:         "invalid json",
			body:         `invalid`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "invalid request",
		},
		{
			name:         "missing email",
			body:         `{"password": "password123"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "email and password are required",
		},
		{
			name:         "missing password",
			body:         `{"email": "test@example.com"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "email and password are required",
		},
		{
			name: "invalid credentials",
			body: `{"email": "test@example.com", "password": "wrongpassword"}`,
			setupService: func(m *ServiceMock) {
				m.ValidatePasswordFunc = func(ctx context.Context, email, password string) (*User, error) {
					return nil, ErrInvalidCreds
				}
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    "invalid credentials",
		},
		{
			name: "service error",
			body: `{"email": "test@example.com", "password": "password123"}`,
			setupService: func(m *ServiceMock) {
				m.ValidatePasswordFunc = func(ctx context.Context, email, password string) (*User, error) {
					return nil, assert.AnError
				}
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &ServiceMock{}
			tt.setupService(mockSvc)

			h := NewHandler(mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.Login(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.checkToken {
				assert.Contains(t, w.Header().Get("Authorization"), "Bearer ")
				var resp AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, testEmail, resp.User.Email)
			}

			if tt.wantErr != "" {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], tt.wantErr)
			}
		})
	}
}
