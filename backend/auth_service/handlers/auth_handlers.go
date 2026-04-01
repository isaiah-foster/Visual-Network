package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-sql-driver/mysql"
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

type createUserRequest struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
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

func (h *AuthHandlers) CreateUser(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, ok := RoleFromContext(request.Context())
	if !ok || role != "admin" {
		http.Error(writer, "forbidden", http.StatusForbidden)
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(writer, "bad request", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		http.Error(writer, "user_id, email, password, and role are required", http.StatusBadRequest)
		return
	}

	passHash, err := auth.HashPasswordArgon2id(req.Password, auth.DefaultArgon2idParameters)
	if err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}

	if err := h.Store.CreateUser(request.Context(), req.UserID, req.Email, passHash, req.Role); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			http.Error(writer, "duplicate user_id", http.StatusConflict)
			return
		}
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
}
