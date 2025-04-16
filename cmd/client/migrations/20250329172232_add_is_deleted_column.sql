-- +goose Up
-- +goose StatementBegin
ALTER TABLE passwords
ADD is_deleted boolean DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE passwords
DROP COLUMN is_deleted;
-- +goose StatementEnd
