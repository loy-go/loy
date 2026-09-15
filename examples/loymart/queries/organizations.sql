-- name: GetOrganizationByID :one
SELECT * FROM organizations
WHERE id = $1 LIMIT 1;

-- name: ListOrganizations :many
SELECT * FROM organizations
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CreateOrganization :one
INSERT INTO organizations (
    name
    , plan
) VALUES (
    $1
    , $2
) RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM organizations
WHERE id = $1;
