-- +goose Up
-- +goose StatementBegin
CREATE TABLE session (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT NOT NULL UNIQUE,
  value TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE session;
-- +goose StatementEnd
