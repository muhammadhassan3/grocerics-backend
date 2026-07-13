-- +goose Up
CREATE TABLE platforms (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code               TEXT NOT NULL,
    display_name       TEXT NOT NULL,
    logo_url           TEXT,
    deep_link_template TEXT,
    delivery_eta_text  TEXT,
    enabled            BOOLEAN NOT NULL DEFAULT true,
    display_order      INT NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ,
    deleted_by         UUID
);
CREATE UNIQUE INDEX idx_platforms_code ON platforms (code) WHERE deleted_at IS NULL;

CREATE TABLE product_platform_links (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id   UUID NOT NULL REFERENCES product_variants (id),
    platform_id  UUID NOT NULL REFERENCES platforms (id),
    platform_sku TEXT,
    product_url  TEXT,
    deep_link    TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    deleted_by   UUID
);
CREATE UNIQUE INDEX idx_product_platform_links_variant_platform
    ON product_platform_links (variant_id, platform_id) WHERE deleted_at IS NULL;

CREATE TABLE platform_prices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id      UUID NOT NULL REFERENCES product_variants (id),
    platform_id     UUID NOT NULL REFERENCES platforms (id),
    city_id         UUID NOT NULL REFERENCES cities (id),
    price_paise     BIGINT NOT NULL,
    mrp_paise       BIGINT,
    available       BOOLEAN NOT NULL DEFAULT true,
    source          VARCHAR NOT NULL DEFAULT 'api',
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_platform_prices_variant_platform_city
    ON platform_prices (variant_id, platform_id, city_id);
CREATE INDEX idx_platform_prices_city_variant ON platform_prices (city_id, variant_id);
CREATE INDEX idx_platform_prices_available
    ON platform_prices (variant_id, city_id) WHERE available;

CREATE TABLE variant_price_summaries (
    variant_id               UUID NOT NULL REFERENCES product_variants (id),
    city_id                  UUID NOT NULL REFERENCES cities (id),
    avg_price_paise          BIGINT,
    min_price_paise          BIGINT,
    min_platform_id          UUID REFERENCES platforms (id),
    available_platform_count INT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (variant_id, city_id)
);

-- +goose Down
DROP TABLE IF EXISTS variant_price_summaries;
DROP TABLE IF EXISTS platform_prices;
DROP TABLE IF EXISTS product_platform_links;
DROP TABLE IF EXISTS platforms;
