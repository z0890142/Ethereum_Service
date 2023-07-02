package consumer

import (
	"Ethereum_Service/internal/data"
	"Ethereum_Service/pkg/model"
	"Ethereum_Service/pkg/utils/logger"
	"context"
	"time"
)

type logConsumer struct {
	logChan         chan model.LogRow
	storeBufferSize int
	storeInterval   time.Duration
	mysqlHandler    *data.MysqlHandler
}

func (c *logConsumer) Run() {
	logBuf := make([]*model.LogRow, 0)
	t := time.NewTicker(c.storeInterval)
	for {
		select {
		case log, ok := <-c.logChan:
			if !ok {
				err := c.mysqlHandler.SaveLogRow(context.Background(), logBuf)
				if err != nil {
					logger.Errorf("LogConsumer Error : %w ", err)
				}
				logBuf = make([]*model.LogRow, 0)
				return
			}
			logBuf = append(logBuf, &log)
			if len(logBuf) == c.storeBufferSize {
				// save
				err := c.mysqlHandler.SaveLogRow(context.Background(), logBuf)
				if err != nil {
					logger.Errorf("LogConsumer Error : %w ", err)
				}
				logBuf = make([]*model.LogRow, 0)
			}
		case <-t.C:
			if len(logBuf) == 0 {
				continue
			}
			// save
			err := c.mysqlHandler.SaveLogRow(context.Background(), logBuf)
			if err != nil {
				logger.Errorf("LogConsumer Error : %w ", err)
			}
			logBuf = make([]*model.LogRow, 0)
		}
	}
}

func (c *logConsumer) GetChan() interface{} {
	return c.logChan
}

func (c *logConsumer) Shutdown() {
	close(c.logChan)
}
