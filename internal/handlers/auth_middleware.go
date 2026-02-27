package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"net/http"

	"userauth/internal/store"
)

type ctxKey string

const (
	SessionCookieName = "session"

	userIDKey ctxKey = "user_id"
	roleKey   ctxKey = "role"
)

func hashSessionToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func WithAuth(s *store.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			http.Error(writer, "unauthorized", http.StatusUnauthorized) //Error code 401
			return
		}

		tokenHash := hashSessionToken(cookie.Value)

		user, err := s.GetUserBySessionHash(request.Context(), tokenHash)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(writer, "unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(writer, "server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(request.Context(), userIDKey, user.UserID)
		ctx = context.WithValue(ctx, roleKey, user.Role)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	value := ctx.Value(userIDKey)
	string, ok := value.(string)
	return string, ok
}

func RoleFromContext(ctx context.Context) (string, bool) {
	value := ctx.Value(roleKey)
	string, ok := value.(string)
	return string, ok
}
