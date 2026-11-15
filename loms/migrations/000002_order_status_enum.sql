-- +goose Up
CREATE TYPE loms.order_status AS ENUM (
    'new',
    'awaiting payment',
    'failed',
    'paid',
    'cancelled'
);

ALTER TABLE loms.orders
    ALTER COLUMN status TYPE loms.order_status
    USING (
        CASE status
            WHEN 'awaiting_payment' THEN 'awaiting payment'
            ELSE status
        END
    )::loms.order_status;

-- +goose Down
ALTER TABLE loms.orders
    ALTER COLUMN status TYPE TEXT
    USING (
        CASE status::TEXT
            WHEN 'awaiting payment' THEN 'awaiting_payment'
            ELSE status::TEXT
        END
    );

DROP TYPE loms.order_status;
