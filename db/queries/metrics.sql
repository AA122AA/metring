-- name: Get :one
SELECT name, type, delta, value, hash FROM metrics
WHERE name = $1 LIMIT 1;

-- name: GetAll :many
SELECT name, type, delta, value, hash FROM metrics
ORDER BY id;

-- name: Write :exec
INSERT INTO metrics (
  name, type, delta, value, hash
) VALUES (
  $1,$2,$3,$4,$5
);

-- name: Update :exec
UPDATE metrics SET (delta, value, updated_at) = ($1, $2, $3)
WHERE name = $4;
