package data

import (
	"Ethereum_Service/c"
	"Ethereum_Service/config"
	"Ethereum_Service/pkg/model"
	"Ethereum_Service/pkg/utils/common"
	"context"
	"fmt"

	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

type MysqlHandler struct {
	gormClient *gorm.DB
}

func NewMysqlHandler(databaseOpts *config.DatabaseOption) (DataHandler, error) {
	db, err := common.OpenMysqlDatabase(databaseOpts)
	if err != nil {
		return nil, fmt.Errorf("NewMysqlHandler: %s", err)
	}
	if err := common.Migration(db); err != nil {
		return nil, fmt.Errorf("NewMysqlHandler: %s", err)
	}

	gormClient, err := gorm.Open(gormMysql.New(gormMysql.Config{
		SkipInitializeWithVersion: true,
		Conn:                      db,
	}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		return nil, fmt.Errorf("NewMysqlHandler: %v", err)
	}

	return &MysqlHandler{
		gormClient: gormClient,
	}, nil
}

func (m *MysqlHandler) UpdateLatestBlockNumber(ctx context.Context, blockNumber int64) error {
	err := m.gormClient.
		Table(c.LatestBlockNumber).
		WithContext(ctx).
		Where("id = ?", 0).
		Update("block_number", blockNumber).Error

	if err != nil {
		return fmt.Errorf("SaveLatestBlockNumber : %w", err)
	}
	return nil
}

func (m *MysqlHandler) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	var blockNumber int64
	err := m.gormClient.
		Table(c.LatestBlockNumber).
		WithContext(ctx).
		Where("id = ?", 0).
		Select("block_number").
		Scan(&blockNumber).Error

	if err != nil {
		return 0, fmt.Errorf("GetLatestBlockNumber : %w", err)
	}
	return blockNumber, nil
}

func (m *MysqlHandler) GetBlockRow(ctx context.Context, blockRow *model.BlockRow) error {
	err := m.gormClient.
		Table(c.Block).
		WithContext(ctx).
		Where(blockRow).
		First(blockRow).Error

	if err != nil {
		return fmt.Errorf("GetBlockRow : %w", err)
	}
	return nil
}

func (m *MysqlHandler) SaveBlockRows(ctx context.Context, blockRow []*model.BlockRow) error {
	err := m.gormClient.Clauses(clause.Insert{Modifier: "IGNORE"}).
		Table(c.Block).WithContext(ctx).Create(blockRow).Error
	if err != nil {
		return fmt.Errorf("SaveBlockRows : %w", err)
	}
	return nil
}

func (m *MysqlHandler) SaveTransactionRow(ctx context.Context, txRow []*model.TransactionRow) error {
	err := m.gormClient.Clauses(clause.Insert{Modifier: "IGNORE"}).
		Table(c.Tx).WithContext(ctx).Create(txRow).Error
	if err != nil {
		return fmt.Errorf("SaveTransactionRow : %w", err)
	}
	return nil
}

func (m *MysqlHandler) SaveLogRow(ctx context.Context, logRow []*model.LogRow) error {
	err := m.gormClient.Clauses(clause.Insert{Modifier: "IGNORE"}).
		Table(c.Log).WithContext(ctx).Create(logRow).Error
	if err != nil {
		return fmt.Errorf("SaveLogRow : %w", err)
	}
	return nil
}

func (m *MysqlHandler) GetBlockRowByBlockNumbers(ctx context.Context, numbers []int64) ([]model.BlockRow, error) {
	var blockRows []model.BlockRow
	err := m.gormClient.
		Table(c.Block).
		WithContext(ctx).
		Where("number IN ?", numbers).
		Find(&blockRows).Error

	if err != nil {
		return nil, fmt.Errorf("GetBlockRowByBlockNumbers : %w", err)
	}
	return blockRows, nil
}

func (m *MysqlHandler) GetLogRowByTxHash(ctx context.Context, txHash string) ([]model.LogRow, error) {
	var logRows []model.LogRow
	err := m.gormClient.
		Table(c.Log).
		WithContext(ctx).
		Where("tx_hash = ?", txHash).
		Find(&logRows).Error

	if err != nil {
		return nil, fmt.Errorf("GetLogRowByTxHash : %w", err)
	}
	return logRows, nil
}

func (m *MysqlHandler) GetTransactionRow(ctx context.Context, tx *model.TransactionRow) error {
	err := m.gormClient.
		Table(c.Tx).
		WithContext(ctx).
		Where(tx).
		First(tx).Error

	if err != nil {
		return fmt.Errorf("GetTransactionRow : %w", err)
	}
	return nil
}

func (m *MysqlHandler) GetTransactionRowByBlockNumber(ctx context.Context, blockNumber int64) ([]model.TransactionRow, error) {
	var txRows []model.TransactionRow
	err := m.gormClient.
		Table(c.Tx).
		WithContext(ctx).
		Where("block_number = ?", blockNumber).
		Find(&txRows).Error

	if err != nil {
		return nil, fmt.Errorf("GetTransactionRowByBlockNumber : %w", err)
	}
	return txRows, nil
}
