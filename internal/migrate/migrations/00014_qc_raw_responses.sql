-- +goose Up
-- Write-only sink: every QuickCommerce HTTP call, captured raw at doGet.
-- Nothing in the app reads this table -- it exists solely to hand the ML
-- person real QC data
CREATE TABLE IF NOT EXISTS qc_raw_responses (
    id            uuid PRIMARY KEY,
    endpoint      text        NOT NULL,
    params        jsonb       NOT NULL DEFAULT '{}'::jsonb,
    status_code   int         NOT NULL,
    response      jsonb,
    response_text text,
    error         text,
    duration_ms   int         NOT NULL DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE qc_raw_responses IS
    'Write-only QuickCommerce call log for ML. Never read by the application.';
COMMENT ON COLUMN qc_raw_responses.response IS
    'Complete unmodified response body. NULL when the body was not valid JSON -- see response_text.';
COMMENT ON COLUMN qc_raw_responses.status_code IS
    '0 when the request never completed (transport error); see error.';

CREATE INDEX IF NOT EXISTS idx_qc_raw_responses_endpoint_created
    ON qc_raw_responses (endpoint, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS qc_raw_responses;
