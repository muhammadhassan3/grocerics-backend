-- +goose Up
CREATE TABLE admins (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name          text        NOT NULL,
    email         text        NOT NULL,
    password_hash text        NOT NULL,
    status        varchar     NOT NULL DEFAULT 'active',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz,
    deleted_by    uuid         
);
CREATE UNIQUE INDEX idx_admins_email ON admins (email) WHERE deleted_at IS NULL;

CREATE TABLE admin_refresh_tokens (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id   uuid        NOT NULL REFERENCES admins (id),
    token_hash varchar(64) NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_admin_refresh_tokens_admin ON admin_refresh_tokens (admin_id);

CREATE TABLE admin_password_resets (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id   uuid        NOT NULL REFERENCES admins (id),
    token_hash varchar(64) NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO admins (id, name, email, password_hash, status, created_at, updated_at, deleted_at)
SELECT id, name, email, password, status, created_at, updated_at, deleted_at
FROM users WHERE role = 'admin';

DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE role = 'admin');
DELETE FROM password_resets WHERE user_id IN (SELECT id FROM users WHERE role = 'admin');
DELETE FROM users WHERE role = 'admin';

ALTER TABLE users DROP COLUMN IF EXISTS email;
ALTER TABLE users DROP COLUMN IF EXISTS password;
ALTER TABLE users DROP COLUMN IF EXISTS role;
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users (phone) WHERE deleted_at IS NULL;

-- password_resets is meaningless for OTP clients; drop it (admins have their own).
DROP TABLE IF EXISTS password_resets;

-- +goose Down
DROP INDEX IF EXISTS idx_users_phone;
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS role varchar NOT NULL DEFAULT 'user';
CREATE TABLE IF NOT EXISTS password_resets (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid        NOT NULL REFERENCES users (id),
    token_hash varchar(64) NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO users (id, name, email, password, role, status, created_at, updated_at, deleted_at)
SELECT id, name, email, password_hash, 'admin', status, created_at, updated_at, deleted_at
FROM admins;
DROP TABLE IF EXISTS admin_password_resets;
DROP TABLE IF EXISTS admin_refresh_tokens;
DROP TABLE IF EXISTS admins;
