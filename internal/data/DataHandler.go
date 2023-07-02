package data

import (
	"Ethereum_Service/pkg/model"
	"context"
)

type DataHandler interface {
	SaveBlockRows(ctx context.Context, blockRow []*model.BlockRow) error
	SaveTransactionRow(ctx context.Context, txRow []*model.TransactionRow) error
	SaveLogRow(ctx context.Context, logRow []*model.LogRow) error

	GetTransactionRow(ctx context.Context, tx *model.TransactionRow) error
	GetLogRowByTxHash(ctx context.Context, txHash string) ([]model.LogRow, error)

	GetBlockRow(ctx context.Context, blockRow *model.BlockRow) error
	GetTransactionRowByBlockNumber(ctx context.Context, blockNumber int64) ([]model.TransactionRow, error)

	GetBlockRowByBlockNumbers(ctx context.Context, numbers []uint64) ([]model.BlockRow, error)

	UpdateLatestBlockNumber(ctx context.Context, blockNumber int64) error
	GetLatestBlockNumber(ctx context.Context) (int64, error)
}
