package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"visual-network/auth_service/auth"
	"visual-network/auth_service/store"
)

type AuthHandlers struct {
	Store *store.Store

	CookieSecure bool

	SessionTTL time.Duration
}

type loginRequest struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

func (h *AuthHandlers) Login(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(writer, "bad request", http.StatusBadRequest)
		return
	}
	if req.UserID == "" || req.Password == "" {
		http.Error(writer, "user_id and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.Store.GetUserForLogin(request.Context(), req.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(writer, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}

	if !user.IsActive {
		http.Error(writer, "account is inactive", http.StatusForbidden)
		return
	}

	ok, err := auth.VerifyPasswordArgon2id(user.PassHash, req.Password)
	if err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(writer, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, tokenHashArr, err := auth.NewSessionToken()
	if err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}

	ttl := h.SessionTTL
	if ttl == 0 {
		ttl = 8 * time.Hour
	}
	expeiresAt := time.Now().Add(ttl)
	tokenHash := tokenHashArr[:]

	if err := h.Store.SetUserSession(request.Context(), user.UserID, tokenHash, expeiresAt); err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expeiresAt,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	writer.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandlers) Logout(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := UserIDFromContext(request.Context())
	if !ok {
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.Store.SetUserSession(request.Context(), userID, nil, time.Now()); err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	writer.WriteHeader(http.StatusNoContent)
}
