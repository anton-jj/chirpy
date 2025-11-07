-- name: CreateUser :one 
INSERT INTO users (id, created_at, updated_at, email, hashed_password) 
VALUES (gen_random_uuid(), NOW(),  NOW(), $1, $2) RETURNING *;


-- name: ResetDatabase :exec
DELETE from users;

-- name: GetUserByEmail :one
	SELECT * FROM users WHERE email = $1;

-- name: GetUserById :one
	SELECT * FROM users WHERE id = $1;
-- name: GetUserFromRefreshToken :one
	SELECT u.* FROM users u JOIN refresh_tokens rt ON u.id = rt.user_id WHERE rt.token = $1;

-- name: UpdateUser :exec
	UPDATE users SET email = $1, hashed_password = $2 WHERE id = $3; 
