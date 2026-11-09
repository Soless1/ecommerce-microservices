-- +goose Up
CREATE SCHEMA IF NOT EXISTS loms;

CREATE TABLE loms.product
(
    sku        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    price      BIGINT      NOT NULL,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE loms.stocks
(
    sku        BIGINT PRIMARY KEY REFERENCES loms.product (sku),
    count      BIGINT      NOT NULL CHECK (count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE loms.orders
(
    order_id   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    status     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE loms.order_items
(
    order_id BIGINT NOT NULL REFERENCES loms.orders (order_id) ON DELETE CASCADE,
    sku      BIGINT NOT NULL REFERENCES loms.product (sku),
    count    BIGINT NOT NULL CHECK (count > 0),
    PRIMARY KEY (order_id, sku)
);

-- +goose Down
DROP TABLE IF EXISTS loms.order_items;
DROP TABLE IF EXISTS loms.orders;
DROP TABLE IF EXISTS loms.stocks;
DROP TABLE IF EXISTS loms.product;
DROP SCHEMA IF EXISTS loms;
