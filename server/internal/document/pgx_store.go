package document

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	FileName    string     `json:"file_name"`
	FilePath    string     `json:"file_path"`
	FileSize    int        `json:"file_size"`
	ContentType string     `json:"content_type"`
	Checksum    string     `json:"checksum"`
	Status      string     `json:"status"`
	UploadedAt  *time.Time `json:"uploaded_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type UpdateRequest struct {
	Status   string
	FileSize int
	Checksum string
}

//go:generate moq -rm -fmt goimports -out pgx_store_mock.go . PgxStore

type PgxStore interface {
	Create(ctx context.Context, fileID, userID uuid.UUID, fileName, filePath, contentType string) (*Document, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Document, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Document, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateRequest) (*Document, error)
}
