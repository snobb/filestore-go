package filestore

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadFileHandler(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func(*FileStoreMock)
		wantStatus int
		wantErr    string
	}{
		{
			name: "success",
			path: "/file_store/uploads/user1/file123_test.pdf",
			body: "file content here",
			setupMock: func(m *FileStoreMock) {
				m.SaveFunc = func(path string, data io.Reader) (*FileInfo, error) {
					return &FileInfo{
						Path:     path,
						Checksum: "abc123",
						Size:     17,
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing path",
			path:       "/file_store/uploads",
			body:       "file content",
			setupMock:  func(m *FileStoreMock) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    "path required",
		},
		{
			name: "save error",
			path: "/file_store/uploads/user1/file123_test.pdf",
			body: "file content",
			setupMock: func(m *FileStoreMock) {
				m.SaveFunc = func(path string, data io.Reader) (*FileInfo, error) {
					return nil, assert.AnError
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &FileStoreMock{}
			tt.setupMock(mockFS)

			h := NewHandler(mockFS)

			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.UploadFileHandler(w, req)

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

func TestDownloadFileHandler(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func(*FileStoreMock)
		wantStatus int
	}{
		{
			name:       "missing path",
			path:       "/file_store/downloads",
			setupMock:  func(m *FileStoreMock) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "disk path returned",
			path: "/file_store/downloads/user1/test.pdf",
			setupMock: func(m *FileStoreMock) {
				m.DiskPathFunc = func(s string) string {
					return "/tmp/" + s
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &FileStoreMock{}
			tt.setupMock(mockFS)

			h := NewHandler(mockFS)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			h.DownloadFileHandler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
