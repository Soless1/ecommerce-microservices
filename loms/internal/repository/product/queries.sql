-- name: CreateProduct :one
INSERT INTO loms.product (name, price)
VALUES (sqlc.arg(name), sqlc.arg(price))
RETURNING sku, name, price;

-- name: GetProduct :one
SELECT sku, name, price
FROM loms.product
WHERE sku = sqlc.arg(sku);
