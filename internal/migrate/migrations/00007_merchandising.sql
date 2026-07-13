-- +goose Up
CREATE TABLE banners (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_url     TEXT NOT NULL,
    target_type   VARCHAR NOT NULL DEFAULT 'none',
    target_id     UUID,
    target_url    TEXT,
    start_date    TIMESTAMPTZ,
    end_date      TIMESTAMPTZ,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    display_order INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    deleted_by    UUID
);
CREATE INDEX idx_banners_active ON banners (is_active, display_order);

CREATE TABLE search_events (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID REFERENCES users (id),
    query             TEXT NOT NULL,
    result_product_id UUID REFERENCES products (id),
    city_id           UUID REFERENCES cities (id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_search_events_created ON search_events (created_at);
CREATE INDEX idx_search_events_query ON search_events (query);

-- +goose Down
DROP TABLE IF EXISTS search_events;
DROP TABLE IF EXISTS banners;
