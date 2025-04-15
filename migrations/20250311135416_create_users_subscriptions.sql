-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(30) NOT NULL,
    last_name VARCHAR(30) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    username VARCHAR(30) UNIQUE,
    password VARCHAR(255) NOT NULL,
    is_resetting_password BOOLEAN DEFAULT FALSE,
    reset_password_token VARCHAR(200) UNIQUE,
    date_reset_password TIMESTAMP,
    profile_picture_url VARCHAR(512),
    stripe_customer_id VARCHAR(255) UNIQUE,
    country VARCHAR(30),
    currency VARCHAR(10) DEFAULT 'EUR',
    refresh_token VARCHAR(512),
    two_fa_enabled BOOLEAN DEFAULT FALSE,
    two_fa_secret VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stripe_customer_id VARCHAR(255) NOT NULL REFERENCES users(stripe_customer_id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) NOT NULL,
    stripe_price_id VARCHAR(255) NOT NULL,
    subscription_type VARCHAR(30) NOT NULL,
    status VARCHAR(30) NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    canceled_at TIMESTAMP,
    ends_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(30) NOT NULL,
    category VARCHAR(30),
    color VARCHAR(20),
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    logo_url VARCHAR(255),
    billing_cycle VARCHAR(20) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE powens_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    powens_id BIGINT NOT NULL UNIQUE,
    access_token VARCHAR(512) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticker VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    sector VARCHAR(50),
    industry VARCHAR(50),
    exchange VARCHAR(50),
    currency VARCHAR(10) DEFAULT 'USD',
    country VARCHAR(50),
    logo_url VARCHAR(255),
    dividend_yield DECIMAL(5,2) CHECK (dividend_yield >= 0),
    pe_ratio DECIMAL(8,2) CHECK (pe_ratio >= 0),
    market_cap BIGINT CHECK (market_cap >= 0), 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stock_id UUID NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    shares DECIMAL(12,6) NOT NULL CHECK (shares >= 0),
    avg_price DECIMAL(10,2) NOT NULL CHECK (avg_price >= 0),
    notes TEXT,
    purchase_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, stock_id)
);
-- Indexes users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_stripe ON users(stripe_customer_id);

-- Indexes subscriptions
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_active ON subscriptions(is_active);
CREATE INDEX idx_subscriptions_user_start ON subscriptions(user_id, start_date);

-- Indexes stocks
CREATE UNIQUE INDEX idx_stocks_ticker ON stocks(ticker);

-- Indexes user_stocks
CREATE INDEX idx_user_stocks_user_id ON user_stocks(user_id);
CREATE UNIQUE INDEX idx_user_stocks_unique ON user_stocks(user_id, stock_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_stocks_unique;
DROP INDEX IF EXISTS idx_user_stocks_user_id;
DROP INDEX IF EXISTS idx_stocks_ticker;
DROP INDEX IF EXISTS idx_subscriptions_user_start;
DROP INDEX IF EXISTS idx_subscriptions_active;
DROP INDEX IF EXISTS idx_subscriptions_user;
DROP INDEX IF EXISTS idx_users_stripe;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS user_stocks;
DROP TABLE IF EXISTS stocks;
DROP TABLE IF EXISTS powens_accounts;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS user_subscriptions;
DROP TABLE IF EXISTS users;