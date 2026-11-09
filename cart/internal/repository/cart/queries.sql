-- name: AddItem :exec
INSERT INTO cart.cart_items (user_id, sku, count)
VALUES (sqlc.arg(user_id), sqlc.arg(sku), sqlc.arg(count))
ON CONFLICT (user_id, sku)
    DO UPDATE SET count      = cart.cart_items.count + EXCLUDED.count,
                  updated_at = NOW();

-- name: DeleteItem :execrows
DELETE
FROM cart.cart_items
WHERE user_id = sqlc.arg(user_id)
  and sku = sqlc.arg(sku);

-- name: GetItemsByUserID :many
SELECT sku, count
FROM cart.cart_items
WHERE user_id = sqlc.arg(user_id)
ORDER BY sku;

-- name: DeleteItemsByUserID :execrows
DELETE
FROM cart.cart_items
WHERE user_id = sqlc.arg(user_id);

