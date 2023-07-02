package consumer

import (
	"Ethereum_Service/c"
	"Ethereum_Service/config"
	"Ethereum_Service/internal/data"
	"Ethereum_Service/pkg/model"
	"time"
)

type Consumer interface {
	Run()
	GetChan() interface{}
	Shutdown()
}

type ConsumerConf struct {
	Type            string
	StoreBufferSize int
	StoreInterval   time.Duration
}

func NewConsumer(conf *ConsumerConf) Consumer {
	mysqlHandler, err := data.NewMysqlHandler(&config.GetConfig().Databases)
	if err != nil {
		panic(err)
	}
	switch conf.Type {
	case c.BlockConsumerType:
		return &blockConsumer{
			blockChan:       make(chan model.BlockRow),
			storeBufferSize: conf.StoreBufferSize,
			storeInterval:   conf.StoreInterval,
			mysqlHandler:    mysqlHandler,
		}
	case c.TxConsumerType:
		return &txConsumer{
			txChan:          make(chan model.TransactionRow),
			storeBufferSize: conf.StoreBufferSize,
			storeInterval:   conf.StoreInterval,
			mysqlHandler:    mysqlHandler,
		}
	case c.LogConsumerType:
		return &logConsumer{
			logChan:         make(chan model.LogRow),
			storeBufferSize: conf.StoreBufferSize,
			storeInterval:   conf.StoreInterval,
			mysqlHandler:    mysqlHandler,
		}
	}
	return nil
}
