package auth

import (
	"context"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

func MockAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// normally in prod: parse jwt and get id.
		// in this poc: get from header or just hardcode.
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			userID = "00000000-0000-0000-0000-000000000000" // Fallback
		}

		// Inject into context
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
