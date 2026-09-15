-- name: GetBillingByID :one
SELECT * FROM billings
WHERE id = $1 LIMIT 1;

-- name: ListBillings :many
SELECT * FROM billings
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CreateBilling :one
INSERT INTO billings (
    amount
    , status
) VALUES (
    $1
    , $2
) RETURNING *;

-- name: DeleteBilling :exec
DELETE FROM billings
WHERE id = $1;
