package filestore

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

const UploadBasePrefix = "/file_store/uploads"
const DownloadBasePrefix = "/file_store/downloads"

type Handler struct {
	filestore FileStore
}

func NewHandler(filestore FileStore) *Handler {
	return &Handler{
		filestore: filestore,
	}
}

// UploadFileHandler saves a file to disk.
//
// input:
//
//	binary data
func (h *Handler) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	pathParam := strings.TrimPrefix(r.URL.Path, UploadBasePrefix)
	if pathParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "path required"})
		return
	}

	info, err := h.filestore.Save(pathParam, r.Body)
	if err != nil {
		slog.Error("failed to save file", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(info)
}

// DownloadFileHandler serves binary file from disk.
func (h *Handler) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	pathParam := strings.TrimPrefix(r.URL.Path, DownloadBasePrefix)
	if pathParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, h.filestore.DiskPath(pathParam))
}
