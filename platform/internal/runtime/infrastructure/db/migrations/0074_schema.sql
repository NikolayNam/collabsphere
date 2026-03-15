-- +goose Up

CREATE SCHEMA IF NOT EXISTS payments;

-- +goose Down

DROP SCHEMA IF EXISTS payments;
