CREATE SCHEMA IF NOT EXISTS "MerchStore";


CREATE TABLE IF NOT EXISTS "MerchStore".users (
    id TEXT PRIMARY KEY,
    username VARCHAR(255) UNIQUE,
    password TEXT NOT NULL,
    balance INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_username_covering ON "MerchStore".users (username) INCLUDE (id, password);
