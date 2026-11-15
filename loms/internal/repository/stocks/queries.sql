-- name: SetStock :exec
INSERT INTO loms.stocks (sku, count)
VALUES (sqlc.arg(sku), sqlc.arg(count))
ON CONFLICT (sku)
DO UPDATE SET count = EXCLUDED.count,
              updated_at = NOW();

-- name: GetStock :one
SELECT count
FROM loms.stocks
WHERE sku = sqlc.arg(sku);

-- name: ReserveStock :execrows
UPDATE loms.stocks
SET count = count - sqlc.arg(count),
    updated_at = NOW()
WHERE sku = sqlc.arg(sku)
  AND count >= sqlc.arg(count);

-- name: ReleaseStock :execrows
UPDATE loms.stocks
SET count = count + sqlc.arg(count),
    updated_at = NOW()
WHERE sku = sqlc.arg(sku);
