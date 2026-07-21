-- +goose Up
ALTER TABLE notification_preferences RENAME COLUMN price_alerts TO deals;
ALTER TABLE notification_preferences DROP COLUMN order_updates;
ALTER TABLE notification_preferences ADD COLUMN muted boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE notification_preferences DROP COLUMN muted;
ALTER TABLE notification_preferences ADD COLUMN order_updates boolean NOT NULL DEFAULT true;
ALTER TABLE notification_preferences RENAME COLUMN deals TO price_alerts;
