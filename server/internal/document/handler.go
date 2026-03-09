package document

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"server/internal/auth"
	"server/internal/filestore"
)

type UploadPendingRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type UploadPendingResponse struct {
	ID        uuid.UUID `json:"id"`
	UploadURL string    `json:"upload_url"`
	StatusURL string    `json:"status_url"`
}

type UpdateStatusRequest struct {
	Status   string `json:"status"`
	FileSize int    `json:"file_size"`
	Checksum string `json:"checksum"`
}

type Handler struct {
	store     PgxStore
	filestore filestore.FileStore
}

func NewHandler(store PgxStore, filestore filestore.FileStore) *Handler {
	return &Handler{
		store:     store,
		filestore: filestore,
	}
}

// UploadPendingHandler - create upload url and save to DB
// POST /api/documents
// input payload:
//
//	{
//	    "file_name": "passport.pdf",
//	    "content_type": "application/pdf"
//	}
//
// output payload:
//
//	{
//	  "id": "uuid-123",
//	  "upload_url": "/file_store/uploads/uuid-123",
//	  "status_url": "/api/documents/uuid-123/status"
//	}
func (h *Handler) UploadPendingHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		// XXX: ignore errors at this point, but normally need to handle.
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized",
		})
		return
	}

	var req UploadPendingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("invalid request", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid request",
		})
		return
	}

	// TODO: improve validation - naive protection against path injection.
	if req.FileName == "" || strings.Contains(req.FileName, "..") {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid file name",
		})
		return
	}

	// TODO: improve validation - content type is required.
	if req.ContentType == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid content type",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("invalid user ID", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid user ID",
		})
		return
	}

	fileID := uuid.New()
	sanitizedFileName := filepath.Base(req.FileName)

	// path to store in DB - /{userID}/{fileID}_{filename}.{ext}
	storePath := filepath.Join(userID,
		fmt.Sprintf("%s_%s", fileID.String(), sanitizedFileName))

	// path to store on disk /{storage-path}/{userID}/{fileID}_{filename}.{ext}
	uploadPath := filepath.Join(filestore.UploadBasePrefix, storePath)

	doc, err := h.store.Create(
		r.Context(),
		fileID,
		userUUID,
		req.FileName,
		storePath,
		req.ContentType,
	)
	if err != nil {
		slog.Error("unable to create document", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unable to write to db > " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(UploadPendingResponse{
		ID:        doc.ID,
		UploadURL: uploadPath,
		StatusURL: fmt.Sprintf("/api/documents/%s/status", doc.ID.String()),
	})
	if err != nil {
		slog.Error("unable to encode response", "error", err.Error())
	}
}

// UpdateDocumentStatusHandler - update document status
// PATCH /api/documents/{id}/status
// input payload:
//
//	{
//		"status": "uploaded"
//		"file_size": 123
//		"checksum": "checksum-123
//	}
//
// output payload:
//
//	{
//	  "id": "uuid-123",
//	  "status": "pending"
//	}
func (h *Handler) UpdateDocumentStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized",
		})
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("invalid id", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		return
	}

	doc, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		slog.Error("unable to load document", "error", err.Error())
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "document not found"})
		return
	}

	if doc.UserID.String() != userID {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "access denied"})
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("unable to decode request", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	updatedDoc, err := h.store.Update(r.Context(), id, &UpdateRequest{
		Status:   req.Status,
		FileSize: req.FileSize,
		Checksum: req.Checksum,
	})
	if err != nil {
		slog.Error("unable to update document", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unable to update document"})
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(updatedDoc); err != nil {
		slog.Error("unable to encode response", "error", err.Error())
	}
}

// GetDocumentHandler - return a document by ID
// GET /api/documents/{id}
//
// output payload:
// document.Document
func (h *Handler) GetDocumentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized",
		})
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		return
	}

	doc, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		slog.Error("unable to load document", "error", err.Error(), "id", id.String())
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "document not found"})
		return
	}

	if doc.UserID.String() != userID {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "access denied"})
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		slog.Error("unable to encode response", "error", err.Error())
	}
}

// ListDocumentsHandler - list files
// GET /api/documents
//
// output payload:
//
//	[]document.Document
func (h *Handler) ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("invalid user ID", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid user ID",
		})
		return
	}

	docs, err := h.store.GetByUserID(r.Context(), userUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("unable to load documents", "error", err.Error())
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unable to load documents",
		})
		return
	}

	if docs == nil {
		docs = []*Document{}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(docs); err != nil {
		slog.Error("unable to encode response", "error", err.Error())
	}
}
