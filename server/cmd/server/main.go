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
	"server/internal/user"
)

func JSONMiddleware(next http.Handler) http.Handler {
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

	pepper := os.Getenv("DATABASE_PEPPER")
	if pepper == "" {
		log.Fatal("CRITICAL ERROR: DATABASE_PEPPER environment variable is not set.")
	}

	jwtKey := os.Getenv("JWT_SECRET")
	if jwtKey == "" {
		log.Fatal("CRITICAL ERROR: JWT_SECRET environment variable is not set.")
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

	userPgxStore := user.NewDefaultPgxStore(dbPool)
	userService := user.NewDefaultService(
		userPgxStore,
		[]byte(pepper),
		0, // use default
		0, // use default
		[]byte(jwtKey),
	)
	userHandlers := user.NewHandler(userService)

	fileStore := filestore.New(fileStoragePath)
	fsHandlers := filestore.NewHandler(fileStore)

	docPgxStore := document.NewDefaultPgxStore(dbPool)
	docService := document.NewService(docPgxStore, fileStore)
	docHandlers := document.NewHandler(docService)

	// JWT middleware for protected routes
	jwtMiddleware := auth.NewJWTMiddleware([]byte(jwtKey))

	// user endpoints (public)
	mux.Handle("POST /api/auth/register",
		JSONMiddleware(http.HandlerFunc(userHandlers.Register)))
	mux.Handle("POST /api/auth/login",
		JSONMiddleware(http.HandlerFunc(userHandlers.Login)))

	// document endpoints (protected)
	mux.Handle("POST /api/documents",
		JSONMiddleware(jwtMiddleware.Middleware(http.HandlerFunc(docHandlers.UploadPendingHandler))))
	mux.Handle("GET /api/documents",
		JSONMiddleware(jwtMiddleware.Middleware(http.HandlerFunc(docHandlers.ListDocumentsHandler))))
	mux.Handle("PATCH /api/documents/{id}/status",
		JSONMiddleware(jwtMiddleware.Middleware(http.HandlerFunc(docHandlers.UpdateDocumentStatusHandler))))
	mux.Handle("GET /api/documents/{id}",
		JSONMiddleware(jwtMiddleware.Middleware(http.HandlerFunc(docHandlers.GetDocumentHandler))))

	// filestore endpoints
	mux.Handle("/file_store/uploads/",
		jwtMiddleware.Middleware(http.HandlerFunc(fsHandlers.UploadFileHandler)))
	mux.Handle("/file_store/downloads/",
		jwtMiddleware.Middleware(http.HandlerFunc(fsHandlers.DownloadFileHandler)))

	// react endpoints - register last as catchall
	registerReactFiles(mux)

	slog.Info("listening on port 3000")
	return http.ListenAndServe(":3000", mux)
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
