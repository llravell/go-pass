-- +goose Up
-- +goose StatementBegin
CREATE TABLE cards (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  cardholder_name TEXT NOT NULL,
  number_encrypted TEXT NOT NULL,
  cvv_encrypted TEXT NOT NULL,
  meta TEXT,
  version INTEGER DEFAULT 0,
  user_id INTEGER NOT NULL,
  is_deleted boolean DEFAULT FALSE,
  expiration_date DATE NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT unique_user_card UNIQUE (user_id, name),
  CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE cards;
-- +goose StatementEnd
