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

-- Add indexes for performance
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes first
DROP INDEX IF EXISTS idx_accounts_user_id;

-- Drop tables in the reverse order of creation to avoid foreign key constraints
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd