-- +goose Up
-- +goose StatementBegin
CREATE TABLE passwords (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  encrypted_pass TEXT NOT NULL,
  meta TEXT,
  version INTEGER DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE passwords;
-- +goose StatementEnd
