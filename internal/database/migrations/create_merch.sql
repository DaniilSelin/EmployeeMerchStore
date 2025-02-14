CREATE TABLE IF NOT EXISTS "MerchStore".merch (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE,
    price INTEGER NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_merch_name ON "MerchStore".merch (name);