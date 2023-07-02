package consumer

import (
	"Ethereum_Service/internal/data"
	"Ethereum_Service/pkg/model"
	"Ethereum_Service/pkg/utils/logger"
	"context"
	"time"
)

type txConsumer struct {
	txChan          chan model.TransactionRow
	storeBufferSize int
	storeInterval   time.Duration
	mysqlHandler    data.DataHandler
}

func (c *txConsumer) Run() {
	txBuf := make([]*model.TransactionRow, 0)
	t := time.NewTicker(c.storeInterval)
	for {
		select {
		case tx, ok := <-c.txChan:
			if !ok {
				// save
				err := c.mysqlHandler.SaveTransactionRow(context.Background(), txBuf)
				if err != nil {
					logger.Errorf("TxConsumer Error : %w ", err)
				}
				txBuf = make([]*model.TransactionRow, 0)
				return
			}
			txBuf = append(txBuf, &tx)
			if len(txBuf) == c.storeBufferSize {
				err := c.mysqlHandler.SaveTransactionRow(context.Background(), txBuf)
				if err != nil {
					logger.Errorf("TxConsumer Error : %w ", err)
				}
				txBuf = make([]*model.TransactionRow, 0)
			}
		case <-t.C:
			if len(txBuf) == 0 {
				continue
			}
			err := c.mysqlHandler.SaveTransactionRow(context.Background(), txBuf)
			if err != nil {
				logger.Errorf("TxConsumer Error : %w ", err)
			}
			txBuf = make([]*model.TransactionRow, 0)
		}
	}
}

func (c *txConsumer) GetChan() interface{} {
	return c.txChan
}

func (c *txConsumer) Shutdown() {
	close(c.txChan)
}
