package indexer_service

import (
	"Ethereum_Service/c"
	"Ethereum_Service/config"
	"Ethereum_Service/internal/consumer"
	"Ethereum_Service/internal/scanner"
	"Ethereum_Service/pkg/model"
	"Ethereum_Service/pkg/utils/logger"
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Service struct {
	ethClient    *ethclient.Client
	blockScanner scanner.BlockScanner
	txScanner    scanner.TxScanner

	latestBlockNumber int64

	blockConsumer consumer.Consumer
	txConsumer    consumer.Consumer
	logConsumer   consumer.Consumer

	shutdownOnce sync.Once
}

func NewService(rcpEndpoint string) (*Service, error) {
	ethClient, err := ethclient.Dial(rcpEndpoint)
	if err != nil {
		return nil, fmt.Errorf("newDefaultEthHandler : %s", err.Error())
	}

	blockScanner := scanner.NewDefaultBlockScanner(ethClient)
	txScanner := scanner.NewDefaultTxScanner(ethClient)

	blockConsume := consumer.NewConsumer(&consumer.ConsumerConf{
		Type:            c.BlockConsumerType,
		StoreBufferSize: config.GetConfig().StoreBufferSize,
		StoreInterval:   time.Duration(10) * time.Second,
	})

	txConsume := consumer.NewConsumer(&consumer.ConsumerConf{
		Type:            c.TxConsumerType,
		StoreBufferSize: config.GetConfig().StoreBufferSize,
		StoreInterval:   time.Duration(10) * time.Second,
	})
	logConsume := consumer.NewConsumer(&consumer.ConsumerConf{
		Type:            c.LogConsumerType,
		StoreBufferSize: config.GetConfig().StoreBufferSize,
		StoreInterval:   time.Duration(10) * time.Second,
	})
	s := Service{
		ethClient:     ethClient,
		blockScanner:  blockScanner,
		txScanner:     txScanner,
		blockConsumer: blockConsume,
		txConsumer:    txConsume,
		logConsumer:   logConsume,
		shutdownOnce:  sync.Once{},
	}

	return &s, nil
}

func (s *Service) Start(workerCount int) {
	logger.GetLogger().Sugar().Infof("start indexer service with %d workers", workerCount)
	ctx := context.Background()

	go s.blockConsumer.Run()
	go s.txConsumer.Run()
	go s.logConsumer.Run()

	latestBlockNumber, err := s.ethClient.BlockNumber(ctx)
	if err != nil {
		panic(err)
	}
	// go s.subscribe()

	interval := latestBlockNumber / uint64(workerCount)
	for i := 0; i < workerCount; i++ {
		logger.GetLogger().Sugar().Infof("start worker %d", i)
		if i == 0 {
			go s.scan(ctx, 0, uint64(i+1)*interval)
			continue
		} else if i == workerCount-1 {
			go s.scan(ctx, uint64(i)*interval+1, latestBlockNumber)
			continue
		}
		go s.scan(ctx, uint64(i)*interval+1, uint64(i+1)*interval)
	}

}

func (s *Service) scan(ctx context.Context, from, to uint64) {
	ctx = context.Background()
	for i := from; i <= to; i++ {
		blockNumber := big.NewInt(int64(i))
		block, err := s.blockScanner.BlockByNumber(ctx, blockNumber)
		if err != nil {
			logger.GetLogger().Sugar().Errorf("scan block %d error: %s", i, err.Error())
			continue
		}
		s.scanBlockInfo(ctx, block)
	}
	logger.GetLogger().Sugar().Infof("scan block from %d to %d done", from, to)
}

func (s *Service) scanBlockInfo(ctx context.Context, block *types.Block) {
	blockRow := convertBlockToRow(block)
	// sned to chan

	s.blockConsumer.GetChan().(chan model.BlockRow) <- blockRow

	for _, tx := range block.Transactions() {
		txRow, err := convertTxToRow(tx, *block.Number())
		if err != nil {
			logger.LoadExtra(map[string]interface{}{
				"err": err.Error(),
			}).Error("convert tx to row error")
			continue
		}
		// sned to chan
		s.txConsumer.GetChan().(chan model.TransactionRow) <- txRow
		logs, err := s.txScanner.LogsByTxHash(ctx, tx.Hash())
		if err != nil && !strings.Contains(err.Error(), "LogsByTxHash : not found") {
			logger.LoadExtra(map[string]interface{}{
				"err": err.Error(),
			}).Error("get logs error")
			continue
		}
		for _, log := range logs {
			logRow := convertLogToRow(log, tx.Hash().Hex())
			// sned to chan
			s.logConsumer.GetChan().(chan model.LogRow) <- logRow
		}
	}

}

func (s *Service) subscribe() {
	headers := make(chan *types.Header)
	ctx := context.Background()
	sub, err := s.ethClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		logger.LoadExtra(map[string]interface{}{
			"err": err.Error(),
		}).Error("subscribe error")
		panic(err)
	}
	for {
		select {
		case err := <-sub.Err():
			logger.LoadExtra(map[string]interface{}{
				"err": err.Error(),
			}).Error("subscribe error")
		case header := <-headers:
			logger.LoadExtra(map[string]interface{}{
				"blockNumber": header.Number.String(),
			}).Info("new block")
			block, err := s.blockScanner.BlockByHash(ctx, header.Hash())
			if err != nil {
				logger.LoadExtra(map[string]interface{}{
					"err": err.Error(),
				}).Error("subscribe error")
				continue
			}
			s.scanBlockInfo(ctx, block)

		}
	}

}

func (s *Service) Shutdown() {

}
