package document

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

	"server/internal/auth"
)

const testUserID = "00000000-0000-0000-0000-000000000000"

func setupTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/documents", h.UploadPendingHandler)
	mux.HandleFunc("GET /api/documents", h.ListDocumentsHandler)
	mux.HandleFunc("PATCH /api/documents/{id}/status", h.UpdateDocumentStatusHandler)
	mux.HandleFunc("GET /api/documents/{id}", h.GetDocumentHandler)
	return mux
}

func TestUploadPendingHandler(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		setupService func(*ServiceMock)
		wantStatus   int
		wantErr      string
		authEnabled  bool
	}{
		{
			name: "success",
			body: `{"file_name": "test.pdf", "content_type": "application/pdf"}`,
			setupService: func(m *ServiceMock) {
				m.UploadPendingFunc = func(ctx context.Context, userID uuid.UUID, fileName, contentType string) (*UploadPendingResponse, error) {
					return &UploadPendingResponse{
						ID:        uuid.New(),
						UploadURL: "/file_store/uploads/test.pdf",
						StatusURL: "/api/documents/uuid/status",
					}, nil
				}
			},
			wantStatus:  http.StatusOK,
			authEnabled: true,
		},
		{
			name:         "unauthorized - no user",
			body:         `{"file_name": "test.pdf", "content_type": "application/pdf"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusUnauthorized,
			wantErr:      "unauthorized",
			authEnabled:  false,
		},
		{
			name:         "invalid json",
			body:         `invalid`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "invalid request",
			authEnabled:  true,
		},
		{
			name:         "empty filename",
			body:         `{"file_name": "", "content_type": "application/pdf"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "invalid file name",
			authEnabled:  true,
		},
		{
			name:         "path traversal in filename",
			body:         `{"file_name": "../etc/passwd", "content_type": "application/pdf"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "invalid file name",
			authEnabled:  true,
		},
		{
			name:         "empty content type",
			body:         `{"file_name": "test.pdf", "content_type": ""}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "invalid content type",
			authEnabled:  true,
		},
		{
			name: "service error",
			body: `{"file_name": "test.pdf", "content_type": "application/pdf"}`,
			setupService: func(m *ServiceMock) {
				m.UploadPendingFunc = func(ctx context.Context, userID uuid.UUID, fileName, contentType string) (*UploadPendingResponse, error) {
					return nil, assert.AnError
				}
			},
			wantStatus:  http.StatusInternalServerError,
			wantErr:     "unable to write to db",
			authEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &ServiceMock{}
			tt.setupService(mockSvc)

			h := NewHandler(mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/documents", strings.NewReader(tt.body))
			if tt.authEnabled {
				req.Header.Set("X-User-ID", testUserID)
			}
			w := httptest.NewRecorder()

			if tt.authEnabled {
				mux := setupTestMux(h)
				handler := auth.MockAuthMiddleware(mux)
				handler.ServeHTTP(w, req)
			} else {
				h.UploadPendingHandler(w, req)
			}

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

func TestUpdateDocumentStatusHandler(t *testing.T) {
	testUserUUID := uuid.MustParse(testUserID)
	docID := uuid.New()

	tests := []struct {
		name         string
		docID        string
		body         string
		setupService func(*ServiceMock)
		wantStatus   int
		wantErr      string
		authEnabled  bool
	}{
		{
			name:  "success",
			docID: docID.String(),
			body:  `{"status": "uploaded", "file_size": 1024, "checksum": "abc123"}`,
			setupService: func(m *ServiceMock) {
				m.GetDocumentFunc = func(ctx context.Context, userID, docID uuid.UUID) (*Document, error) {
					return &Document{
						ID:         docID,
						UserID:     testUserUUID,
						FileName:   "test.pdf",
						FilePath:   "/path/test.pdf",
						Status:     "pending",
						UploadedAt: nil,
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}, nil
				}
				m.UpdateStatusFunc = func(ctx context.Context, userID, docID uuid.UUID, status string, fileSize int, checksum string) (*Document, error) {
					return &Document{
						ID:         docID,
						UserID:     testUserUUID,
						FileName:   "test.pdf",
						Status:     status,
						FileSize:   fileSize,
						Checksum:   checksum,
						UploadedAt: nil,
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}, nil
				}
			},
			wantStatus:  http.StatusOK,
			authEnabled: true,
		},
		{
			name:         "unauthorized - no user",
			docID:        docID.String(),
			body:         `{"status": "uploaded"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusUnauthorized,
			wantErr:      "unauthorized",
			authEnabled:  false,
		},
		{
			name:         "invalid id format",
			docID:        "invalid",
			body:         `{"status": "uploaded"}`,
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "invalid id",
			authEnabled:  true,
		},
		{
			name:  "document not found",
			docID: docID.String(),
			body:  `{"status": "uploaded"}`,
			setupService: func(m *ServiceMock) {
				m.GetDocumentFunc = func(ctx context.Context, userID, docID uuid.UUID) (*Document, error) {
					return nil, assert.AnError
				}
			},
			wantStatus:  http.StatusNotFound,
			wantErr:     "document not found",
			authEnabled: true,
		},
		{
			name:  "forbidden - wrong user",
			docID: docID.String(),
			body:  `{"status": "uploaded"}`,
			setupService: func(m *ServiceMock) {
				m.GetDocumentFunc = func(ctx context.Context, userID, docID uuid.UUID) (*Document, error) {
					return &Document{
						ID:         docID,
						UserID:     uuid.New(),
						FileName:   "test.pdf",
						FilePath:   "/path/test.pdf",
						Status:     "pending",
						UploadedAt: nil,
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}, nil
				}
			},
			wantStatus:  http.StatusForbidden,
			wantErr:     "access denied",
			authEnabled: true,
		},
		{
			name:  "invalid request body",
			docID: docID.String(),
			body:  `invalid`,
			setupService: func(m *ServiceMock) {
				m.GetDocumentFunc = func(ctx context.Context, userID, docID uuid.UUID) (*Document, error) {
					return &Document{
						ID:         docID,
						UserID:     testUserUUID,
						FileName:   "test.pdf",
						FilePath:   "/path/test.pdf",
						Status:     "pending",
						UploadedAt: nil,
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}, nil
				}
			},
			wantStatus:  http.StatusBadRequest,
			wantErr:     "invalid request",
			authEnabled: true,
		},
		{
			name:  "service error",
			docID: docID.String(),
			body:  `{"status": "uploaded"}`,
			setupService: func(m *ServiceMock) {
				m.GetDocumentFunc = func(ctx context.Context, userID, docID uuid.UUID) (*Document, error) {
					return &Document{
						ID:         docID,
						UserID:     testUserUUID,
						FileName:   "test.pdf",
						FilePath:   "/path/test.pdf",
						Status:     "pending",
						UploadedAt: nil,
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}, nil
				}
				m.UpdateStatusFunc = func(ctx context.Context, userID, docID uuid.UUID, status string, fileSize int, checksum string) (*Document, error) {
					return nil, assert.AnError
				}
			},
			wantStatus:  http.StatusInternalServerError,
			wantErr:     "unable to update document",
			authEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &ServiceMock{}
			tt.setupService(mockSvc)

			h := NewHandler(mockSvc)

			req := httptest.NewRequest(http.MethodPatch, "/api/documents/"+tt.docID+"/status", strings.NewReader(tt.body))
			if tt.authEnabled && tt.docID != "invalid" {
				req.Header.Set("X-User-ID", testUserID)
			}
			w := httptest.NewRecorder()

			if tt.authEnabled {
				mux := setupTestMux(h)
				handler := auth.MockAuthMiddleware(mux)
				handler.ServeHTTP(w, req)
			} else {
				h.UpdateDocumentStatusHandler(w, req)
			}

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

func TestGetDocumentHandler(t *testing.T) {
	testUserUUID := uuid.MustParse(testUserID)
	docID := uuid.New()

	tests := []struct {
		name         string
		docID        string
		setupService func(*ServiceMock)
		wantStatus   int
		wantErr      string
		authEnabled  bool
	}{
		{
			name:  "success",
			docID: docID.String(),
			setupService: func(m *ServiceMock) {
				m.GetDocumentFunc = func(ctx context.Context, userID, docID uuid.UUID) (*Document, error) {
					return &Document{
						ID:         docID,
						UserID:     testUserUUID,
						FileName:   "test.pdf",
						FilePath:   "/path/test.pdf",
						Status:     "pending",
						UploadedAt: nil,
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}, nil
				}
			},
			wantStatus:  http.StatusOK,
			authEnabled: true,
		},
		{
			name:         "unauthorized - no user",
			docID:        docID.String(),
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusUnauthorized,
			wantErr:      "unauthorized",
			authEnabled:  false,
		},
		{
			name:         "invalid id format",
			docID:        "invalid",
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "invalid id",
			authEnabled:  true,
		},
		{
			name:  "document not found",
			docID: docID.String(),
			setupService: func(m *ServiceMock) {
				m.GetDocumentFunc = func(ctx context.Context, userID, docID uuid.UUID) (*Document, error) {
					return nil, assert.AnError
				}
			},
			wantStatus:  http.StatusNotFound,
			wantErr:     "document not found",
			authEnabled: true,
		},
		{
			name:  "forbidden - wrong user",
			docID: docID.String(),
			setupService: func(m *ServiceMock) {
				m.GetDocumentFunc = func(ctx context.Context, userID, docID uuid.UUID) (*Document, error) {
					return &Document{
						ID:         docID,
						UserID:     uuid.New(),
						FileName:   "test.pdf",
						FilePath:   "/path/test.pdf",
						Status:     "pending",
						UploadedAt: nil,
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}, nil
				}
			},
			wantStatus:  http.StatusForbidden,
			wantErr:     "access denied",
			authEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &ServiceMock{}
			tt.setupService(mockSvc)

			h := NewHandler(mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/api/documents/"+tt.docID, nil)
			if tt.authEnabled && tt.docID != "invalid" {
				req.Header.Set("X-User-ID", testUserID)
			}
			w := httptest.NewRecorder()

			if tt.authEnabled {
				mux := setupTestMux(h)
				handler := auth.MockAuthMiddleware(mux)
				handler.ServeHTTP(w, req)
			} else {
				h.GetDocumentHandler(w, req)
			}

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

func TestListDocumentsHandler(t *testing.T) {
	tests := []struct {
		name         string
		setupService func(*ServiceMock)
		wantStatus   int
		wantErr      string
		authEnabled  bool
	}{
		{
			name: "success",
			setupService: func(m *ServiceMock) {
				m.ListDocumentsFunc = func(ctx context.Context, userID uuid.UUID) ([]*Document, error) {
					return []*Document{
						{
							ID:         uuid.New(),
							UserID:     userID,
							FileName:   "test1.pdf",
							FilePath:   "/path/test1.pdf",
							Status:     "pending",
							UploadedAt: nil,
							UpdatedAt:  time.Now(),
							CreatedAt:  time.Now(),
						},
						{
							ID:         uuid.New(),
							UserID:     userID,
							FileName:   "test2.pdf",
							FilePath:   "/path/test2.pdf",
							Status:     "verified",
							UploadedAt: nil,
							UpdatedAt:  time.Now(),
							CreatedAt:  time.Now(),
						},
					}, nil
				}
			},
			wantStatus:  http.StatusOK,
			authEnabled: true,
		},
		{
			name:         "unauthorized - no user",
			setupService: func(m *ServiceMock) {},
			wantStatus:   http.StatusUnauthorized,
			wantErr:      "unauthorized",
			authEnabled:  false,
		},
		{
			name: "success - empty list",
			setupService: func(m *ServiceMock) {
				m.ListDocumentsFunc = func(ctx context.Context, userID uuid.UUID) ([]*Document, error) {
					return []*Document{}, nil
				}
			},
			wantStatus:  http.StatusOK,
			authEnabled: true,
		},
		{
			name: "service error",
			setupService: func(m *ServiceMock) {
				m.ListDocumentsFunc = func(ctx context.Context, userID uuid.UUID) ([]*Document, error) {
					return nil, assert.AnError
				}
			},
			wantStatus:  http.StatusInternalServerError,
			wantErr:     "unable to load documents",
			authEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &ServiceMock{}
			tt.setupService(mockSvc)

			h := NewHandler(mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
			if tt.authEnabled {
				req.Header.Set("X-User-ID", testUserID)
			}
			w := httptest.NewRecorder()

			if tt.authEnabled {
				mux := setupTestMux(h)
				handler := auth.MockAuthMiddleware(mux)
				handler.ServeHTTP(w, req)
			} else {
				h.ListDocumentsHandler(w, req)
			}

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

func TestHandler_UploadPendingHandler_AuthValidation(t *testing.T) {
	h := NewHandler(&ServiceMock{})

	req := httptest.NewRequest(http.MethodPost, "/api/documents", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	h.UploadPendingHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_GetDocumentHandler_AuthValidation(t *testing.T) {
	h := NewHandler(&ServiceMock{})

	req := httptest.NewRequest(http.MethodGet, "/api/documents/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	h.GetDocumentHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_ListDocumentsHandler_AuthValidation(t *testing.T) {
	h := NewHandler(&ServiceMock{})

	req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	w := httptest.NewRecorder()
	h.ListDocumentsHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_UpdateDocumentStatusHandler_AuthValidation(t *testing.T) {
	h := NewHandler(&ServiceMock{})

	req := httptest.NewRequest(http.MethodPatch, "/api/documents/"+uuid.New().String()+"/status", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	h.UpdateDocumentStatusHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
