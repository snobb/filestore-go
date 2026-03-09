package pglib

import "github.com/jackc/pgx/v5"

//go:generate moq -rm -fmt goimports -out rows_mock.go . Rows
type Rows interface {
	pgx.Rows
}
