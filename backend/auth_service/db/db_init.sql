CREATE DATABASE IF NOT EXISTS visual_network;
USE visual_network;

CREATE TABLE IF NOT EXISTS user_auth (
    user_id VARCHAR(64) NOT NULL,
    email VARCHAR(255) NOT NULL,
    pass_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'user',
    session_token_hash BINARY(32) NULL,
    session_expiration DATETIME NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id),
    UNIQUE KEY uq_user_auth_email (email),
    KEY idx_user_auth_session_token_hash (session_token_hash),
    KEY idx_user_auth_session_expiration (session_expiration)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;