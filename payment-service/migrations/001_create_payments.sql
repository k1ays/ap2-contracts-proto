CREATE TABLE IF NOT EXISTS payments (
    id             VARCHAR(64)  PRIMARY KEY,
    order_id       VARCHAR(64)  NOT NULL,
    transaction_id VARCHAR(128) NOT NULL DEFAULT '',
    amount         BIGINT       NOT NULL,
    status         VARCHAR(32)  NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
