-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS caption
(
    id         UUID PRIMARY KEY     DEFAULT GEN_RANDOM_UUID(),
    text       TEXT        NOT NULL UNIQUE,
    author_id  BIGINT      NOT NULL,
    approved   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS caption;
-- +goose StatementEnd
