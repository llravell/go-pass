-- +goose Up
-- +goose StatementBegin
CREATE TYPE upload_status AS ENUM ('pending', 'done');

CREATE TABLE files (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  size BIGINT DEFAULT 0,
  minio_bucket TEXT NOT NULL,
  meta TEXT,
  user_id INTEGER NOT NULL,
  upload_status upload_status NOT NULL,
  is_deleted boolean DEFAULT FALSE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT unique_user_bucket_file UNIQUE (user_id, minio_bucket, name),
  CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE files;
DROP TYPE upload_status;
-- +goose StatementEnd
