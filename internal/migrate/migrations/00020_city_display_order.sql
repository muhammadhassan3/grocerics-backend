-- +goose Up
ALTER TABLE cities ADD COLUMN display_order integer NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE cities DROP COLUMN display_order;
