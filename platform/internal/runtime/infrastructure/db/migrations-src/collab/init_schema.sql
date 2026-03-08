-- +goose Up

CREATE SCHEMA IF NOT EXISTS collab;

-- +goose Down

DROP SCHEMA collab;
