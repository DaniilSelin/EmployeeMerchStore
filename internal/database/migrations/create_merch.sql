CREATE TABLE IF NOT EXISTS "MerchStore".merch (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE,
    price INTEGER NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_merch_name_price ON "MerchStore".merch (name) INCLUDE (price);

INSERT INTO "MerchStore".merch (name, price, description)
VALUES 
    ('T-Shirt', 80, 'A cool t-shirt'),
    ('cup', 20, 'A nice cup'),
    ('book', 50, 'An interesting book'),
    ('pen', 10, 'A pen'),
    ('powerbank', 200, 'A powerbank'),
    ('hoody', 300, 'A hoody'),
    ('umbrella', 200, 'An umbrella'),
    ('socks', 10, 'Socks'),
    ('wallet', 50, 'A wallet'),
    ('pink-hoody', 500, 'A pink hoody')
ON CONFLICT (name) DO NOTHING;
