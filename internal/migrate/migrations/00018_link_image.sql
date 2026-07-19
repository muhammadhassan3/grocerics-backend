-- +goose Up
-- Per-(variant, platform) image captured from the QuickCommerce search result
-- at link time. Variant search serves the image of the top-ranked linked
-- platform instead of the product-level image.
ALTER TABLE product_platform_links ADD COLUMN image_url text;

-- +goose Down
ALTER TABLE product_platform_links DROP COLUMN image_url;
