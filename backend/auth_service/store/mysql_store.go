package store

import (
	"context"
	"database/sql"
	"time"
)

type Store struct {
	DB *sql.DB
}

type User struct {
	UserID   string
	Email    string
	PassHash string
	Role     string
	IsActive bool
}

func (s *Store) GetUserForLogin(ctx context.Context, userID string) (*User, error) {
	const query = `
	SELECT user_id, email, pass_hash, role, is_active
	FROM user_auth
	WHERE user_id = ?
	LIMIT 1;`

	user := &User{}
	var isActiveInt int
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Email,
		&user.Role,
		&isActiveInt,
	)
	if err != nil {
		return nil, err
	}
	user.IsActive = isActiveInt == 1
	return user, nil
}

func (s *Store) CreateUser(ctx context.Context, userID, email, passHash, role string) error {
	const query = `
	INSERT INTO user_auth (user_id, email, pass_hash, role, is_active)
	VALUES (?, ?, ?, ?, 1);`
	_, err := s.DB.ExecContext(ctx, query, userID, email, passHash, role)
	return err
}

func (s *Store) SetUserSession(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) error {
	const query = `
	UPDATE user_auth
	SET session_token_hash = ?, session_expiration = ?
	WHERE user_id = ? AND is_active = 1;`

	_, err := s.DB.ExecContext(ctx, query, tokenHash, expiresAt, userID)
	return err
}

func (s *Store) ClearUserSession(ctx context.Context, userID string) error {
	const query = `
	UPDATE user_auth
	SET session_token_hash = NULL, session_expiration = NULL
	WHERE user_id = ?;`
	_, err := s.DB.ExecContext(ctx, query, userID)
	return err
}

func (s *Store) GetUserBySessionHash(ctx context.Context, tokenHash []byte) (*User, error) {
	const query = `
	SELECT user_id, email, pass_hash, role, is_active
	FROM user_auth
	WHERE session_token_hash = ?
	AND session_expiration > NOW()
	AND is_active = 1
	LIMIT 1;`

	user := &User{}
	var isActiveInt int
	err := s.DB.QueryRowContext(ctx, query, tokenHash).Scan(
		&user.UserID,
		&user.Email,
		&user.PassHash,
		&user.Role,
		&isActiveInt,
	)
	if err != nil {
		return nil, err
	}
	user.IsActive = isActiveInt == 1
	return user, nil
}
