CREATE OR REPLACE FUNCTION notify_order_updates() RETURNS TRIGGER AS $$
DECLARE
    payload JSON;
BEGIN
    payload := json_build_object(
        'id', NEW.id,
        'customer_id', NEW.customer_id,
        'item_name', NEW.item_name,
        'amount', NEW.amount,
        'status', NEW.status,
        'created_at', NEW.created_at
    );

    PERFORM pg_notify('order_updates', payload::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_order_updates ON orders;

CREATE TRIGGER trg_order_updates
AFTER INSERT OR UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION notify_order_updates();
