package document

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"server/internal/pglib"
)

const (
	StatusPending  = "pending"
	StatusUploaded = "uploaded"
	StatusVerified = "verified"
	StatusRejected = "rejected"
)

type DefaultPgxStore struct {
	db pglib.Pool
}

func NewDefaultPgxStore(db pglib.Pool) *DefaultPgxStore {
	return &DefaultPgxStore{db: db}
}

func (s *DefaultPgxStore) Create(ctx context.Context, fileID, userID uuid.UUID, fileName, filePath, contentType string) (*Document, error) {
	var doc Document
	var pgDocID pgtype.UUID
	err := s.db.QueryRow(ctx, `
		INSERT INTO documents (id, user_id, file_name, file_path, content_type, file_size)
		VALUES ($1, $2, $3, $4, $5, 0)
		ON CONFLICT (user_id, file_name)
		DO UPDATE SET
			id=EXCLUDED.id,
			file_path=EXCLUDED.file_path,
			content_type=EXCLUDED.content_type,
			updated_at = NOW()
		RETURNING id, file_name, file_path, file_size, content_type,
			uploaded_at, updated_at, created_at
	`, fileID, userID, fileName, filePath, contentType).Scan(
		&pgDocID, &doc.FileName, &doc.FilePath,
		&doc.FileSize, &doc.ContentType,
		&doc.UploadedAt, &doc.UpdatedAt, &doc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	docID, err := pglib.ParsePgUUID(pgDocID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID %v > %w", pgDocID, err)
	}

	doc.UserID = userID
	doc.ID = docID

	return &doc, nil
}

func (s *DefaultPgxStore) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	var doc Document
	var pgDocID, pgUserID pgtype.UUID
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, file_name, file_path, file_size, content_type, COALESCE(checksum, '') as checksum, status, uploaded_at, updated_at
		FROM documents WHERE id = $1
	`, id).Scan(
		&pgDocID, &pgUserID, &doc.FileName, &doc.FilePath, &doc.FileSize,
		&doc.ContentType, &doc.Checksum, &doc.Status,
		&doc.UploadedAt, &doc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	docID, err := pglib.ParsePgUUID(pgDocID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID > %w", err)
	}

	userID, err := pglib.ParsePgUUID(pgUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID  > %w", err)
	}

	doc.ID = docID
	doc.UserID = userID

	return &doc, nil
}

func (s *DefaultPgxStore) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Document, error) {
	// XXX: sorting by the upload date but in future may parameterise this.
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, file_name, file_path, file_size, content_type,
			COALESCE(checksum, '') as checksum, status, uploaded_at, updated_at, created_at
		FROM documents WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := []*Document{}
	for rows.Next() {
		var doc Document
		var pgDocID, pgUserID pgtype.UUID
		if err := rows.Scan(
			&pgDocID, &pgUserID, &doc.FileName, &doc.FilePath, &doc.FileSize,
			&doc.ContentType, &doc.Checksum, &doc.Status,
			&doc.UploadedAt, &doc.UpdatedAt, &doc.CreatedAt,
		); err != nil {
			return nil, err
		}

		docID, err := pglib.ParsePgUUID(pgDocID)
		if err != nil {
			return nil, fmt.Errorf("invalid document ID %v > %w", pgDocID, err)
		}

		userID, err := pglib.ParsePgUUID(pgUserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID %v > %w", pgUserID, err)
		}

		doc.ID = docID
		doc.UserID = userID
		docs = append(docs, &doc)
	}
	return docs, rows.Err()
}

func (s *DefaultPgxStore) Update(ctx context.Context, id uuid.UUID, updateRequest *UpdateRequest) (*Document, error) {
	query := `UPDATE documents SET updated_at = NOW()`
	args := []any{}
	argNum := 1

	if updateRequest.Status != "" {
		query += fmt.Sprintf(", status = $%d", argNum)
		args = append(args, updateRequest.Status)
		argNum++
	}

	if updateRequest.FileSize > 0 {
		query += fmt.Sprintf(", file_size = $%d", argNum)
		args = append(args, updateRequest.FileSize)
		argNum++
	}

	if updateRequest.Checksum != "" {
		query += fmt.Sprintf(", checksum = $%d, uploaded_at=NOW()", argNum)
		args = append(args, updateRequest.Checksum)
		argNum++
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, user_id, file_name, file_path, file_size, "+
		"content_type, COALESCE(checksum, '') as checksum, status, uploaded_at, updated_at, created_at", argNum)
	args = append(args, id)

	var doc Document
	var pgDocID, pgUserID pgtype.UUID
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&pgDocID, &pgUserID, &doc.FileName, &doc.FilePath, &doc.FileSize,
		&doc.ContentType, &doc.Checksum, &doc.Status,
		&doc.UploadedAt, &doc.UpdatedAt, &doc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	docID, err := pglib.ParsePgUUID(pgDocID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID %v > %w", pgDocID, err)
	}

	userID, err := pglib.ParsePgUUID(pgUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID %v > %w", pgUserID, err)
	}

	doc.ID = docID
	doc.UserID = userID
	return &doc, nil
}
