-- +goose Up
-- +goose StatementBegin
ALTER TABLE metrics
    ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE metrics
    DROP COLUMN created_at,
    DROP COLUMN updated_at;
-- +goose StatementEnd
