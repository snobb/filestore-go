package document

import (
	"context"
	"path/filepath"
	"server/internal/filestore"

	"github.com/google/uuid"
)

type DefaultService struct {
	store     PgxStore
	filestore filestore.FileStore
}

func NewService(store PgxStore, filestore filestore.FileStore) Service {
	return &DefaultService{
		store:     store,
		filestore: filestore,
	}
}

func (s *DefaultService) UploadPending(ctx context.Context, userID uuid.UUID, fileName, contentType string) (*UploadPendingResponse, error) {
	fileID := uuid.New()
	sanitizedFileName := filepath.Base(fileName)

	storePath := filepath.Join(userID.String(), fileID.String()+"_"+sanitizedFileName)
	uploadPath := filepath.Join(filestore.UploadBasePrefix, storePath)

	doc, err := s.store.Create(ctx, fileID, userID, fileName, storePath, contentType)
	if err != nil {
		return nil, err
	}

	return &UploadPendingResponse{
		ID:        doc.ID,
		UploadURL: uploadPath,
		StatusURL: "/api/documents/" + doc.ID.String() + "/status",
	}, nil
}

func (s *DefaultService) GetDocument(ctx context.Context, userID, docID uuid.UUID) (*Document, error) {
	return s.store.GetByID(ctx, docID)
}

func (s *DefaultService) ListDocuments(ctx context.Context, userID uuid.UUID) ([]*Document, error) {
	docs, err := s.store.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if docs == nil {
		return []*Document{}, nil
	}
	return docs, nil
}

func (s *DefaultService) UpdateStatus(ctx context.Context, userID, docID uuid.UUID, status string, fileSize int, checksum string) (*Document, error) {
	return s.store.Update(ctx, docID, &UpdateRequest{
		Status:   status,
		FileSize: fileSize,
		Checksum: checksum,
	})
}
