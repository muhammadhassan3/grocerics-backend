-- +goose Up
-- City is now derived from an address's lat/lng (nearest enabled city), not a
-- pincode lookup table. The pincodes table had no writer and only fed the old
-- resolveCity(pincode) path, which is gone.
DROP TABLE IF EXISTS pincodes;

-- +goose Down
-- Irreversible: pincodes was never populated by the app and the resolver that
-- read it has been removed. Restore from 00002/00009 if you truly need it back.
