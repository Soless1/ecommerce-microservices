-- +goose Up
CREATE SCHEMA IF NOT EXISTS cart;


CREATE TABLE IF NOT EXISTS cart.cart_items
(
    user_id    BIGINT      NOT NULL,
    sku        BIGINT      NOT NULL,
    count      BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, sku)
);

-- +goose Down
DROP TABLE IF EXISTS cart.cart_items;
DROP SCHEMA IF EXISTS cart;
