-- +goose Up
ALTER TABLE users ADD hashed_password TEXT;

UPDATE users SET hashed_password = '';

ALTER TABLE users ALTER COLUMN hashed_password SET NOT NULL;

ALTER TABLE users ALTER COLUMN hashed_password SET DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS hashed_password;
