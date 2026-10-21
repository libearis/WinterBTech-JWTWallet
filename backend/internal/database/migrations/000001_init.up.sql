CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(254) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(LOWER(email));

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL CHECK (amount > 0),
    notes VARCHAR(140) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_sender ON transactions(sender_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_recipient ON transactions(recipient_id, created_at DESC);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    idempotency_key VARCHAR(128) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    response_code INT NOT NULL,
    response_body JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_created_at ON idempotency_keys(created_at);

-- Seed default demo users and wallets
INSERT INTO users (id, email, password_hash)
VALUES 
    ('11111111-1111-1111-1111-111111111111', 'jane@example.com', '$2a$12$5kD2XmD/UpvojR4xKIh5LOLAdrKP4.fTlnSUq5AF5MXdD7furHemm'),
    ('22222222-2222-2222-2222-222222222222', 'john@example.com', '$2a$12$5kD2XmD/UpvojR4xKIh5LOLAdrKP4.fTlnSUq5AF5MXdD7furHemm')
ON CONFLICT (email) DO NOTHING;

INSERT INTO wallets (user_id, balance, currency)
VALUES
    ('11111111-1111-1111-1111-111111111111', 100000, 'IDR'),
    ('22222222-2222-2222-2222-222222222222', 100000, 'IDR')
ON CONFLICT (user_id) DO NOTHING;

