package data

import (
	"Ethereum_Service/c"
	"Ethereum_Service/config"
	"Ethereum_Service/pkg/model"
	"Ethereum_Service/pkg/utils/common"
	"context"
	"fmt"

	gormMysql "gorm.io/driver/mysql"

	"gorm.io/gorm"
)

type MysqlHandler struct {
	gormClient *gorm.DB
}

func NewMysqlHandler(databaseOpts *config.DatabaseOption) (*MysqlHandler, error) {
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

func (m *MysqlHandler) SaveBlockRows(ctx context.Context, blockRow []*model.BlockRow) error {
	err := m.gormClient.
		Table(c.Block).WithContext(ctx).Create(blockRow).Error
	if err != nil {
		return fmt.Errorf("SaveBlockRows : %w", err)
	}
	return nil
}

func (m *MysqlHandler) SaveTransactionRow(ctx context.Context, txRow []*model.TransactionRow) error {
	err := m.gormClient.
		Table(c.Tx).WithContext(ctx).Create(txRow).Error
	if err != nil {
		return fmt.Errorf("SaveTransactionRow : %w", err)
	}
	return nil
}

func (m *MysqlHandler) SaveLogRow(ctx context.Context, logRow []*model.LogRow) error {
	err := m.gormClient.
		Table(c.Log).WithContext(ctx).Create(logRow).Error
	if err != nil {
		return fmt.Errorf("SaveLogRow : %w", err)
	}
	return nil
}
