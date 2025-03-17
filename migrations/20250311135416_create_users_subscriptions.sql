-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(30) NOT NULL,
    last_name VARCHAR(30) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    username VARCHAR(30) UNIQUE,
    password TEXT NOT NULL,
    is_resetting_password BOOLEAN DEFAULT FALSE,
    reset_password_token VARCHAR(200),
    date_reset_password DATE,
    profile_picture_url VARCHAR(512),
    stripe_customer_id VARCHAR(255) UNIQUE,
    subscription VARCHAR(30),
    bio TEXT,
    country VARCHAR(30),
    last_login TIMESTAMP,
    two_fa_enabled BOOLEAN DEFAULT FALSE,
    two_fa_code VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(30) NOT NULL,
    category VARCHAR(30),
    color VARCHAR(20),
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    logo_url VARCHAR(255),
    is_recuring BOOLEAN DEFAULT TRUE,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE powens_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,  
    powens_id INT NOT NULL UNIQUE,
    access_token VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_stripe ON users(stripe_customer_id);
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_powens_access_token ON powens_accounts(access_token);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS powens_accounts;
