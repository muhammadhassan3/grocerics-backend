-- +goose Up
CREATE TABLE user_addresses (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id),
    label      TEXT,
    line1      TEXT NOT NULL,
    line2      TEXT,
    pincode    TEXT NOT NULL,
    city_id    UUID REFERENCES cities (id),
    lat        DOUBLE PRECISION,
    lng        DOUBLE PRECISION,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    deleted_by UUID
);
CREATE INDEX idx_user_addresses_user_id ON user_addresses (user_id);

CREATE TABLE fcm_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id),
    token      TEXT NOT NULL,
    platform   VARCHAR NOT NULL DEFAULT 'android',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    deleted_by UUID
);
CREATE UNIQUE INDEX idx_fcm_tokens_token ON fcm_tokens (token) WHERE deleted_at IS NULL;

CREATE TABLE notification_preferences (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users (id),
    price_alerts  BOOLEAN NOT NULL DEFAULT true,
    promotions    BOOLEAN NOT NULL DEFAULT true,
    order_updates BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    deleted_by    UUID
);
CREATE UNIQUE INDEX idx_notification_preferences_user_id ON notification_preferences (user_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS fcm_tokens;
DROP TABLE IF EXISTS user_addresses;
