-- name: Get :one
SELECT * FROM metrics
WHERE name = $1 LIMIT 1;

-- name: GetAll :many
SELECT * FROM metrics
ORDER BY id;

-- name: Write :exec
INSERT INTO metrics (
  name, type, delta, value, hash
) VALUES (
  $1,$2,$3,$4,$5
);

-- name: Update :exec
UPDATE metrics SET (delta, value) = ($1, $2)
WHERE name = $3;
