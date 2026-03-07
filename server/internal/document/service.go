package document

import (
	"context"

	"github.com/google/uuid"
)

//go:generate moq -rm -fmt goimports -out service_mock.go . Service

type Service interface {
	UploadPending(ctx context.Context, userID uuid.UUID, fileName, contentType string) (*UploadPendingResponse, error)
	GetDocument(ctx context.Context, userID, docID uuid.UUID) (*Document, error)
	ListDocuments(ctx context.Context, userID uuid.UUID) ([]*Document, error)
	UpdateStatus(ctx context.Context, userID, docID uuid.UUID, status string, fileSize int, checksum string) (*Document, error)
}
