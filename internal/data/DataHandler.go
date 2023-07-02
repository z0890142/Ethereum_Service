package data

import (
	"Ethereum_Service/pkg/model"
	"context"
)

type DataHandler interface {
	SaveBlockRows(ctx context.Context, blockRow []*model.BlockRow) error
	SaveTransactionRow(ctx context.Context, txRow []*model.TransactionRow) error
	SaveLogRow(ctx context.Context, logRow []*model.LogRow) error
	UpdateLatestBlockNumber(ctx context.Context, blockNumber int64) error
	GetLatestBlockNumber(ctx context.Context) (int64, error)
}
