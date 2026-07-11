-- +goose Up
CREATE TABLE categories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL,
    image_url       TEXT,
    description     TEXT,
    is_top_category BOOLEAN NOT NULL DEFAULT false,
    status          VARCHAR NOT NULL DEFAULT 'active',
    display_order   INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    deleted_by      UUID
);
CREATE UNIQUE INDEX idx_categories_slug ON categories (slug) WHERE deleted_at IS NULL;

CREATE TABLE subcategories (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id        UUID NOT NULL REFERENCES categories (id),
    name               TEXT NOT NULL,
    slug               TEXT,
    image_url          TEXT,
    is_top_subcategory BOOLEAN NOT NULL DEFAULT false,
    status             VARCHAR NOT NULL DEFAULT 'active',
    display_order      INT NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ,
    deleted_by         UUID
);
CREATE INDEX idx_subcategories_category_id ON subcategories (category_id);

CREATE TABLE brands (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    slug         TEXT,
    image_url    TEXT,
    is_top_brand BOOLEAN NOT NULL DEFAULT false,
    status       VARCHAR NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    deleted_by   UUID
);

CREATE TABLE products (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id    UUID NOT NULL REFERENCES categories (id),
    subcategory_id UUID REFERENCES subcategories (id),
    brand_id       UUID REFERENCES brands (id),
    name           TEXT NOT NULL,
    description    TEXT,
    image_url      TEXT,
    is_top_item    BOOLEAN NOT NULL DEFAULT false,
    status         VARCHAR NOT NULL DEFAULT 'active',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    deleted_by     UUID
);
CREATE INDEX idx_products_category_id ON products (category_id);

CREATE INDEX idx_products_search ON products
    USING GIN (to_tsvector('simple', name || ' ' || coalesce(description, '')));

CREATE TABLE product_images (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id    UUID NOT NULL REFERENCES products (id),
    image_url     TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    deleted_by    UUID
);
CREATE INDEX idx_product_images_product_id ON product_images (product_id);

CREATE TABLE product_variants (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id        UUID NOT NULL REFERENCES products (id),
    custom_variant_id TEXT,
    volume_value      NUMERIC NOT NULL,
    volume_unit       VARCHAR NOT NULL,
    display_order     INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,
    deleted_by        UUID
);
CREATE INDEX idx_product_variants_product_id ON product_variants (product_id);

-- +goose Down
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS brands;
DROP TABLE IF EXISTS subcategories;
DROP TABLE IF EXISTS categories;
