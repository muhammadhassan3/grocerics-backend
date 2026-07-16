-- +goose Up
ALTER TABLE platform_prices ADD COLUMN inventory INT;

ALTER TABLE platforms ADD COLUMN qc_name TEXT;

ALTER TABLE cities ADD COLUMN default_pincode TEXT;

CREATE TABLE user_activity_daily (
    user_id       UUID NOT NULL,
    activity_date DATE NOT NULL,
    PRIMARY KEY (user_id, activity_date)
);
CREATE INDEX idx_user_activity_daily_date ON user_activity_daily (activity_date);

-- +goose Down
DROP TABLE IF EXISTS user_activity_daily;
ALTER TABLE cities DROP COLUMN IF EXISTS default_pincode;
ALTER TABLE platforms DROP COLUMN IF EXISTS qc_name;
ALTER TABLE platform_prices DROP COLUMN IF EXISTS inventory;
