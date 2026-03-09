package filestore

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type DefaultFileStore struct {
	baseDir string
}

func New(baseDir string) *DefaultFileStore {
	return &DefaultFileStore{
		baseDir: baseDir,
	}
}

func (s *DefaultFileStore) DiskPath(path string) string {
	return filepath.Join(s.baseDir, path)
}

// Save now returns metadata for the DB
func (s *DefaultFileStore) Save(path string, data io.Reader) (*FileInfo, error) {
	slog.Debug("saving file", "baseDir", s.baseDir, "path", path)

	pathOnDisk := filepath.Join(s.baseDir, path)
	dirPath := filepath.Dir(pathOnDisk)

	if err := os.MkdirAll(dirPath, 0750); err != nil && !os.IsExist(err) {
		slog.Error("failed to create directory", "error", err.Error())
		return nil, err
	}

	f, err := os.Create(pathOnDisk)
	if err != nil {
		slog.Error("failed to create file", "error", err.Error())
		return nil, err
	}
	defer f.Close()

	hash := sha256.New()
	mw := io.MultiWriter(f, hash)

	size, err := io.Copy(mw, data)
	if err != nil {
		slog.Error("failed to copy data", "error", err.Error())
		return nil, err
	}

	slog.Debug("file saved", "size", size)
	return &FileInfo{
		Path:     path,
		Checksum: hex.EncodeToString(hash.Sum(nil)),
		Size:     size,
	}, nil
}
