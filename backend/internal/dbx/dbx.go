// Package dbx holds small Postgres helpers shared by the stores.
package dbx

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolation = "23505" // Postgres error code

// UniqueViolation reports whether err is a broken unique rule, and which
// index or constraint it was.
func UniqueViolation(err error) (constraint string, ok bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return pgErr.ConstraintName, true
	}
	return "", false
}
