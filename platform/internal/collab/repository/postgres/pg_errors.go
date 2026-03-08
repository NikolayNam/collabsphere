package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code == pgUniqueViolation
	}
	return false
}

func isForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code == pgForeignKeyViolation
	}
	return false
}

func IsUnique(err error) bool {
	return isUniqueViolation(err)
}

func IsForeignKey(err error) bool {
	return isForeignKeyViolation(err)
}
