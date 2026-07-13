-- +goose Up
CREATE TABLE platform_delivery_etas (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_id     UUID NOT NULL REFERENCES platforms (id),
    pincode         TEXT NOT NULL,
    eta_minutes     INT,
    serviceable     BOOLEAN NOT NULL DEFAULT true,
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_platform_delivery_etas_platform_pincode
    ON platform_delivery_etas (platform_id, pincode);
CREATE INDEX idx_platform_delivery_etas_pincode ON platform_delivery_etas (pincode);

-- +goose Down
DROP TABLE IF EXISTS platform_delivery_etas;
