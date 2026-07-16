-- +goose Up
-- domain.Product declared display_order but no migration ever created the column,
-- so product reorder failed with "column does not exist". Add it.
ALTER TABLE products ADD COLUMN IF NOT EXISTS display_order INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE products DROP COLUMN IF EXISTS display_order;
