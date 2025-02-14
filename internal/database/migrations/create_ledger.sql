CREATE TABLE IF NOT EXISTS "MerchStore".ledger (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    movement_type VARCHAR(50) NOT NULL, -- 'transfer_in', 'transfer_out', 'purchase'
    amount DECIMAL(18, 2) NOT NULL,
    reference_id INTEGER, 
    created_at TIMESTAMP DEFAULT now(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES "MerchStore".users(id)
);