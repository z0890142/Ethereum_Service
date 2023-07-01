package consumer

import (
	"Ethereum_Service/internal/data"
	"Ethereum_Service/pkg/model"
	"Ethereum_Service/pkg/utils/logger"
	"context"
	"time"
)

type blockConsumer struct {
	blockChan       chan model.BlockRow
	storeBufferSize int
	storeInterval   time.Duration
	mysqlHandler    *data.MysqlHandler
}

func (c *blockConsumer) Run() {
	blockBuf := make([]*model.BlockRow, 0)
	t := time.NewTicker(c.storeInterval)
	for {
		select {
		case block, ok := <-c.blockChan:
			if !ok {
				// save
				err := c.mysqlHandler.SaveBlockRows(context.Background(), blockBuf)
				if err != nil {
					logger.Errorf("BlockConsumer Error : %w ", err)
				}
				blockBuf = make([]*model.BlockRow, 0)
				return
			}
			blockBuf = append(blockBuf, &block)
			if len(blockBuf) == c.storeBufferSize {
				// save
				err := c.mysqlHandler.SaveBlockRows(context.Background(), blockBuf)
				if err != nil {
					logger.Errorf("BlockConsumer Error : %w ", err)
				}
				blockBuf = make([]*model.BlockRow, 0)
			}
		case <-t.C:
			if len(blockBuf) == 0 {
				continue
			}
			// save
			err := c.mysqlHandler.SaveBlockRows(context.Background(), blockBuf)
			if err != nil {
				logger.Errorf("BlockConsumer Error : %w ", err)
			}
			blockBuf = make([]*model.BlockRow, 0)
		}
	}
}

func (c *blockConsumer) GetChan() interface{} {
	return c.blockChan
}
