CREATE SCHEMA IF NOT EXISTS "MerchStore";


CREATE TABLE IF NOT EXISTS "MerchStore".users (
    id TEXT PRIMARY KEY,
    username VARCHAR(255) UNIQUE,
    password TEXT NOT NULL,
    balance DECIMAL(18, 2) NOT NULL DEFAULT 1000,
    created_at TIMESTAMP DEFAULT now()
);