-- +goose Up
-- +goose StatementBegin
CREATE TABLE metrics (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  delta BIGINT NOT NULL,
  value DOUBLE PRECISION NOT NULL,
  hash TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE metrics;
-- +goose StatementEnd
