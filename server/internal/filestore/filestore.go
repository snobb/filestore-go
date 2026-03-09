package filestore

import (
	"io"
)

type FileInfo struct {
	Path     string `json:"path"`
	Checksum string `json:"check_sum"`
	Size     int64  `json:"file_size"`
}

//go:generate moq -rm -fmt goimports -out filestore_mock.go . FileStore

type FileStore interface {
	// DiskPath returns the path to the file on disk
	DiskPath(string) string

	// Save now returns metadata for the DB
	Save(path string, data io.Reader) (*FileInfo, error)
}
