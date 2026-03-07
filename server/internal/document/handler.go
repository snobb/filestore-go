package document

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"server/internal/auth"
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
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
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

	if req.FileName == "" || strings.Contains(req.FileName, "..") {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid file name",
		})
		return
	}

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

	resp, err := h.service.UploadPending(
		r.Context(),
		userUUID,
		req.FileName,
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
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

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("invalid user ID", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	doc, err := h.service.GetDocument(r.Context(), userUUID, id)
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

	updatedDoc, err := h.service.UpdateStatus(
		r.Context(),
		userUUID,
		id,
		req.Status,
		req.FileSize,
		req.Checksum,
	)
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

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	doc, err := h.service.GetDocument(r.Context(), userUUID, id)
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

	docs, err := h.service.ListDocuments(r.Context(), userUUID)
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
