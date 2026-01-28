-- +goose Up
-- +goose StatementBegin
ALTER TABLE metrics
    ALTER COLUMN delta DROP NOT NULL,
    ALTER COLUMN value DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE metrics
    ALTER COLUMN delta SET NOT NULL,
    ALTER COLUMN value SET NOT NULL;
-- +goose StatementEnd
