package pglib

import "github.com/jackc/pgx/v5"

//go:generate moq -rm -fmt goimports -out tx_mock.go . Tx

type TxIsoLevel string

const (
	IsoLevelSerializable   = TxIsoLevel(pgx.Serializable)
	IsoLevelRepeatableRead = TxIsoLevel(pgx.RepeatableRead)
	IsoLevelReadCommitted  = TxIsoLevel(pgx.ReadCommitted)
	IsoLevelReadUncommited = TxIsoLevel(pgx.ReadUncommitted)
)

type Tx interface {
	pgx.Tx
}
