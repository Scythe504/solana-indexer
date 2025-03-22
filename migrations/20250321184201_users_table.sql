-- +goose Up
-- +goose StatementBegin
-- Create users table
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    image VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Create accounts table
CREATE TABLE accounts (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    provider_type VARCHAR(255) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    provider_account_id VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255),
    access_token VARCHAR(255),
    access_token_expires TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(provider_account_id)
);

-- Create sessions table
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    expires TIMESTAMP NOT NULL,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    access_token VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create verification_requests table
CREATE TABLE verification_requests (
    id VARCHAR(255) PRIMARY KEY,
    identifier VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Add indexes for performance
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_verification_requests_token ON verification_requests(token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes first
DROP INDEX IF EXISTS idx_verification_requests_token;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_accounts_user_id;

-- Drop tables in the reverse order of creation to avoid foreign key constraints
DROP TABLE IF EXISTS verification_requests;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd