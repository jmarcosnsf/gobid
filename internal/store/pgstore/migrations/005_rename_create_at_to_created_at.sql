-- Write your migrate up statements here

ALTER TABLE bids RENAME COLUMN create_at TO created_at;
---- create above / drop below ----
ALTER TABLE bids RENAME COLUMN created_at TO create_at;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
