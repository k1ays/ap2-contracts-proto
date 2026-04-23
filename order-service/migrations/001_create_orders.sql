CREATE TABLE IF NOT EXISTS orders (
    id          VARCHAR(64)  PRIMARY KEY,
    customer_id VARCHAR(128) NOT NULL,
    item_name   VARCHAR(255) NOT NULL,
    amount      BIGINT       NOT NULL,
    status      VARCHAR(32)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
