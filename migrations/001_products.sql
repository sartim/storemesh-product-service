CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price_minor BIGINT NOT NULL CHECK (price_minor >= 0),
    currency TEXT NOT NULL,
    status SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS products_status_idx ON products (status);
