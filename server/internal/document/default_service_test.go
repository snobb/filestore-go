package document

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDefaultService_UploadPending(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		fileName       string
		contentType    string
		setupStoreMock func(*PgxStoreMock)
		wantErr        bool
	}{
		{
			name:        "success",
			fileName:    "test.pdf",
			contentType: "application/pdf",
			setupStoreMock: func(m *PgxStoreMock) {
				m.CreateFunc = func(ctx context.Context, fileID, userID uuid.UUID, fileName, filePath, contentType string) (*Document, error) {
					return &Document{
						ID:          fileID,
						UserID:      userID,
						FileName:    fileName,
						FilePath:    filePath,
						ContentType: contentType,
						Status:      "pending",
						UploadedAt:  nil,
						UpdatedAt:   time.Now(),
						CreatedAt:   time.Now(),
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:        "store error",
			fileName:    "test.pdf",
			contentType: "application/pdf",
			setupStoreMock: func(m *PgxStoreMock) {
				m.CreateFunc = func(ctx context.Context, fileID, userID uuid.UUID, fileName, filePath, contentType string) (*Document, error) {
					return nil, assert.AnError
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &PgxStoreMock{}
			tt.setupStoreMock(mockStore)

			svc := NewService(mockStore, nil)

			resp, err := svc.UploadPending(context.Background(), userID, tt.fileName, tt.contentType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.ID)
				assert.NotEmpty(t, resp.UploadURL)
				assert.NotEmpty(t, resp.StatusURL)
				assert.Contains(t, resp.StatusURL, "/api/documents/")
				assert.Contains(t, resp.StatusURL, "/status")
			}
		})
	}
}

func TestDefaultService_GetDocument(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	tests := []struct {
		name           string
		setupStoreMock func(*PgxStoreMock)
		want           *Document
		wantErr        bool
	}{
		{
			name: "success",
			setupStoreMock: func(m *PgxStoreMock) {
				m.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*Document, error) {
					return &Document{
						ID:         id,
						UserID:     userID,
						FileName:   "test.pdf",
						FilePath:   "/path/test.pdf",
						Status:     "pending",
						UploadedAt: nil,
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}, nil
				}
			},
			want: &Document{
				ID:       docID,
				UserID:   userID,
				FileName: "test.pdf",
			},
			wantErr: false,
		},
		{
			name: "not found",
			setupStoreMock: func(m *PgxStoreMock) {
				m.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*Document, error) {
					return nil, assert.AnError
				}
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &PgxStoreMock{}
			tt.setupStoreMock(mockStore)

			svc := NewService(mockStore, nil)

			resp, err := svc.GetDocument(context.Background(), userID, docID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, resp.ID)
				assert.Equal(t, tt.want.UserID, resp.UserID)
				assert.Equal(t, tt.want.FileName, resp.FileName)
			}
		})
	}
}

func TestDefaultService_ListDocuments(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		setupStoreMock func(*PgxStoreMock)
		want           []*Document
		wantErr        bool
	}{
		{
			name: "success",
			setupStoreMock: func(m *PgxStoreMock) {
				m.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) ([]*Document, error) {
					return []*Document{
						{ID: uuid.New(), UserID: uid, FileName: "test1.pdf"},
						{ID: uuid.New(), UserID: uid, FileName: "test2.pdf"},
					}, nil
				}
			},
			want: []*Document{
				{UserID: userID, FileName: "test1.pdf"},
				{UserID: userID, FileName: "test2.pdf"},
			},
			wantErr: false,
		},
		{
			name: "empty list",
			setupStoreMock: func(m *PgxStoreMock) {
				m.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) ([]*Document, error) {
					return []*Document{}, nil
				}
			},
			want:    []*Document{},
			wantErr: false,
		},
		{
			name: "nil returns empty slice",
			setupStoreMock: func(m *PgxStoreMock) {
				m.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) ([]*Document, error) {
					return nil, nil
				}
			},
			want:    []*Document{},
			wantErr: false,
		},
		{
			name: "error",
			setupStoreMock: func(m *PgxStoreMock) {
				m.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) ([]*Document, error) {
					return nil, assert.AnError
				}
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &PgxStoreMock{}
			tt.setupStoreMock(mockStore)

			svc := NewService(mockStore, nil)

			resp, err := svc.ListDocuments(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.want), len(resp))
			}
		})
	}
}

func TestDefaultService_UpdateStatus(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	tests := []struct {
		name           string
		status         string
		fileSize       int
		checksum       string
		setupStoreMock func(*PgxStoreMock)
		want           *Document
		wantErr        bool
	}{
		{
			name:     "success",
			status:   "uploaded",
			fileSize: 1024,
			checksum: "abc123",
			setupStoreMock: func(m *PgxStoreMock) {
				m.UpdateFunc = func(ctx context.Context, id uuid.UUID, req *UpdateRequest) (*Document, error) {
					return &Document{
						ID:         id,
						UserID:     userID,
						Status:     req.Status,
						FileSize:   req.FileSize,
						Checksum:   req.Checksum,
						UploadedAt: nil,
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}, nil
				}
			},
			want: &Document{
				ID:       docID,
				UserID:   userID,
				Status:   "uploaded",
				FileSize: 1024,
				Checksum: "abc123",
			},
			wantErr: false,
		},
		{
			name:     "error",
			status:   "uploaded",
			fileSize: 1024,
			checksum: "abc123",
			setupStoreMock: func(m *PgxStoreMock) {
				m.UpdateFunc = func(ctx context.Context, id uuid.UUID, req *UpdateRequest) (*Document, error) {
					return nil, assert.AnError
				}
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &PgxStoreMock{}
			tt.setupStoreMock(mockStore)

			svc := NewService(mockStore, nil)

			resp, err := svc.UpdateStatus(context.Background(), userID, docID, tt.status, tt.fileSize, tt.checksum)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Status, resp.Status)
				assert.Equal(t, tt.want.FileSize, resp.FileSize)
				assert.Equal(t, tt.want.Checksum, resp.Checksum)
			}
		})
	}
}
