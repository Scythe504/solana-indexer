-- +goose Up
-- +goose StatementBegin
-- Create user_database_credentials table
CREATE TABLE user_database_credentials (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    db_name VARCHAR(255),
    host VARCHAR(255),
    "user" VARCHAR(255),
    port INTEGER,
    password VARCHAR(255),
    ssl_mode VARCHAR(50),
    connection_string TEXT,
    connection_limit SMALLINT,
    status VARCHAR(50),
    last_connected_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    error_message TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create address_registry table
CREATE TABLE address_registry (
    id VARCHAR(255) PRIMARY KEY,
    token_address VARCHAR(255) NOT NULL UNIQUE,
    token_name VARCHAR(255) NOT NULL,
    token_symbol VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_fetched_at TIMESTAMP
);

-- Create subscriptions table
CREATE TABLE subscriptions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    token_address VARCHAR(255) NOT NULL,
    indexing_strategy TEXT[] NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    status BOOLEAN NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (token_address) REFERENCES address_registry(token_address)
);

-- Create subscription_lookup table
CREATE TABLE subscription_lookup (
    id VARCHAR(255) PRIMARY KEY,
    token_address VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    strategy VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    helius_webhook_id VARCHAR(255),
    last_updated TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (token_address) REFERENCES address_registry(token_address)
);

-- Create helius_webhook_configs table
CREATE TABLE helius_webhook_configs (
    id VARCHAR(255) PRIMARY KEY,
    webhook_name VARCHAR(255) NOT NULL,
    webhook_id VARCHAR(255) NOT NULL UNIQUE,
    address_count INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Add indexes for performance
CREATE INDEX idx_user_database_credentials_user_id ON user_database_credentials(user_id);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_token_address ON subscriptions(token_address);
CREATE INDEX idx_subscription_lookup_user_id ON subscription_lookup(user_id);
CREATE INDEX idx_subscription_lookup_token_address ON subscription_lookup(token_address);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes first
DROP INDEX IF EXISTS idx_subscription_lookup_token_address;
DROP INDEX IF EXISTS idx_subscription_lookup_user_id;
DROP INDEX IF EXISTS idx_subscriptions_token_address;
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP INDEX IF EXISTS idx_user_database_credentials_user_id;

-- Drop tables in the reverse order of creation to avoid foreign key constraints
DROP TABLE IF EXISTS helius_webhook_configs;
DROP TABLE IF EXISTS subscription_lookup;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS address_registry;
DROP TABLE IF EXISTS user_database_credentials;
-- +goose StatementEnd