package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"server/internal/auth"
	"server/internal/document"
	"server/internal/filestore"
	"server/internal/pglib"
)

func JSONMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Enable debug logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hostname := os.Getenv("POSTGRES_HOST")
	if hostname == "" {
		hostname = "localhost"
	}

	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	if fileStoragePath == "" {
		fileStoragePath = "/filestore"
	}

	dbConfig := pglib.Config{
		Host:     hostname,
		Port:     5432,
		Username: "postgres",
		Password: "postgres",
		Database: "filestore",
	}

	dbPool, err := pglib.NewPool(ctx, dbConfig)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	pgxStore := document.NewDefaultPgxStore(dbPool)
	fileStore := filestore.New(fileStoragePath)

	docHandlers := document.NewHandler(pgxStore, fileStore)
	fsHandlers := filestore.NewHandler(fileStore)

	// document endpoints
	mux.HandleFunc("POST /api/documents",
		JSONMiddleware(docHandlers.UploadPendingHandler))
	mux.HandleFunc("GET /api/documents",
		JSONMiddleware(docHandlers.ListDocumentsHandler))
	mux.HandleFunc("PATCH /api/documents/{id}/status",
		JSONMiddleware(docHandlers.UpdateDocumentStatusHandler))
	mux.HandleFunc("GET /api/documents/{id}",
		JSONMiddleware(docHandlers.GetDocumentHandler))

	// filestore endpoints
	mux.HandleFunc("/file_store/uploads/", fsHandlers.UploadFileHandler)
	mux.HandleFunc("/file_store/downloads/", fsHandlers.DownloadFileHandler)

	// react endpoints - register last as catchall
	registerReactFiles(mux)

	slog.Info("listening on port 3000")
	return http.ListenAndServe(":3000", auth.MockAuthMiddleware(mux))
}

func registerReactFiles(mux *http.ServeMux) {
	distPath := "/app/client/dist"
	fs := http.FileServer(http.Dir(distPath))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("catchall", "path", r.URL.Path)
		path := filepath.Join(distPath, r.URL.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}
