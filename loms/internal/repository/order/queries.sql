-- name: CreateOrder :one
INSERT INTO loms.orders (user_id, status)
VALUES (sqlc.arg(user_id), sqlc.arg(status))
RETURNING order_id, user_id, status, created_at, updated_at;

-- name: CreateOrderItem :exec
INSERT INTO loms.order_items (order_id, sku, count)
VALUES (sqlc.arg(order_id), sqlc.arg(sku), sqlc.arg(count));

-- name: GetOrder :one
SELECT order_id, user_id, status, created_at, updated_at
FROM loms.orders
WHERE order_id = sqlc.arg(order_id);

-- name: GetOrderForUpdate :one
SELECT order_id, user_id, status, created_at, updated_at
FROM loms.orders
WHERE order_id = sqlc.arg(order_id)
FOR UPDATE;

-- name: GetOrderItems :many
SELECT sku, count
FROM loms.order_items
WHERE order_id = sqlc.arg(order_id)
ORDER BY sku;

-- name: SetOrderStatus :execrows
UPDATE loms.orders
SET status = sqlc.arg(status),
    updated_at = NOW()
WHERE order_id = sqlc.arg(order_id);

