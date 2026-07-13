-- +goose Up
CREATE TABLE cities (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    slug          TEXT NOT NULL,
    state         TEXT,
    enabled       BOOLEAN NOT NULL DEFAULT true,
    display_order INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    deleted_by    UUID
);
CREATE UNIQUE INDEX idx_cities_slug ON cities (slug) WHERE deleted_at IS NULL;

CREATE TABLE pincodes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pincode      TEXT NOT NULL,
    city_id      UUID NOT NULL REFERENCES cities (id),
    serviceable  BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    deleted_by   UUID
);
CREATE UNIQUE INDEX idx_pincodes_pincode ON pincodes (pincode) WHERE deleted_at IS NULL;
CREATE INDEX idx_pincodes_city_id ON pincodes (city_id);

ALTER TABLE users ADD CONSTRAINT fk_users_current_city
    FOREIGN KEY (current_city_id) REFERENCES cities (id);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_current_city;
DROP TABLE IF EXISTS pincodes;
DROP TABLE IF EXISTS cities;
