-- +goose Up
-- +goose StatementBegin
CREATE TABLE passwords (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  encrypted_pass TEXT NOT NULL,
  meta TEXT,
  version INTEGER DEFAULT 0,
  user_id INTEGER NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (user_id, name),
  CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE passwords;
-- +goose StatementEnd
