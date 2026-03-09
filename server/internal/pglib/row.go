package pglib

import "github.com/jackc/pgx/v5"

//go:generate moq -rm -fmt goimports -out row_mock.go . Row
type Row interface {
	pgx.Row
}
