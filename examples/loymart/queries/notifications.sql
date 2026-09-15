-- name: GetNotificationByID :one
SELECT * FROM notifications
WHERE id = $1 LIMIT 1;

-- name: ListNotifications :many
SELECT * FROM notifications
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CreateNotification :one
INSERT INTO notifications (
    channel
    , recipient
    , body
) VALUES (
    $1
    , $2
    , $3
) RETURNING *;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1;
