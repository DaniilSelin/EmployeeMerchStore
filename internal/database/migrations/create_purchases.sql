CREATE TABLE IF NOT EXISTS "MerchStore".purchases (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    merch_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    purchased_at TIMESTAMP DEFAULT now(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES "MerchStore".users(id),
    CONSTRAINT fk_merch FOREIGN KEY (merch_id) REFERENCES "MerchStore".merch(id),
    CONSTRAINT uq_user_merch UNIQUE (user_id, merch_id)
);
