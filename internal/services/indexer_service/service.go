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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/streadway/amqp"
)

type Service struct {
	ethClient    *ethclient.Client
	blockScanner scanner.BlockScanner
	txScanner    scanner.TxScanner

	blockConsumer consumer.Consumer
	txConsumer    consumer.Consumer
	logConsumer   consumer.Consumer

	latestBlockNumber int64

	jobChan chan model.Job
	jobDone chan chan model.JobResult
	mqConn  *amqp.Connection

	rcpEndpoint string

	shutDownCtx  context.Context
	cancel       context.CancelFunc
	shutdownOnce sync.Once
}

func NewService(rcpEndpoint string) (*Service, error) {

	ethClient, err := ethclient.Dial(rcpEndpoint)
	if err != nil {
		return nil, fmt.Errorf("newDefaultEthHandler : %s", err.Error())
	}

	blockScanner := scanner.NewDefaultBlockScanner(rcpEndpoint)
	txScanner := scanner.NewDefaultTxScanner(rcpEndpoint)

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
		rcpEndpoint:   rcpEndpoint,
		jobChan:       make(chan model.Job, 1000),
		jobDone:       make(chan chan model.JobResult, 1000),
	}

	return &s, nil
}

func (s *Service) Start(workerCount int) {
	logger.GetLogger().Sugar().Infof("start indexer service with %d workers", workerCount)
	s.shutDownCtx, s.cancel = context.WithCancel(context.Background())

	s.createMqConn()

	go s.blockConsumer.Run()
	go s.txConsumer.Run()
	go s.logConsumer.Run()

	for i := 0; i < workerCount; i++ {
		go s.scan(s.shutDownCtx)
	}
	s.createMqConn()

}

func (s *Service) createMqConn() {
	var err error
	s.mqConn, err = amqp.Dial(config.GetConfig().MQEndpoint)
	if err != nil {
		panic(err)
	}
	go s.startMqConsumer()
	go s.jobDoneHandler(s.shutDownCtx)
}

func (s *Service) startMqConsumer() {
	ch, err := s.mqConn.Channel()
	if err != nil {
		logger.GetLogger().Sugar().Errorf("Failed to open a channel: %v", err)
		return
	}
	defer ch.Close()

	err = ch.Qos(1, 0, false)
	if err != nil {
		logger.GetLogger().Sugar().Errorf("Failed to set prefetch count: %v", err)
		return
	}

	queue, err := ch.QueueDeclare(
		"blockNumber_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.GetLogger().Sugar().Errorf("Failed to declare a queue: %v", err)
		return
	}

	msgs, err := ch.Consume(
		queue.Name,
		"indexer_service",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logger.GetLogger().Sugar().Errorf("Failed to declare a queue: %v", err)
		return
	}

	for msg := range msgs {
		blockNumber, err := strconv.ParseUint(string(msg.Body), 10, 64)
		logger.GetLogger().Sugar().Infof("Received a message: %s", msg.Body)

		if err != nil {
			logger.GetLogger().Sugar().Errorf("Failed to parse block number: %v", err)
			continue
		}
		blockNumberBig := big.NewInt(int64(blockNumber))
		responseChan := make(chan model.JobResult)

		s.jobChan <- model.Job{
			BlockNumber: blockNumberBig,
			DoneChan:    responseChan,
			Msg:         &msg,
		}
		s.jobDone <- responseChan
	}
}

func (s *Service) jobDoneHandler(ctx context.Context) {

	for {
		select {
		case job := <-s.jobDone:
			result := <-job
			if result.BlockNumber == nil {
				logger.GetLogger().Sugar().Errorf("Failed to get block number: %v", result.BlockNumber)
				continue
			}
			result.Msg.Ack(true)
			s.publishBlockDone(result.BlockNumber)

		case <-ctx.Done():
			return
		}
	}

}

func (s *Service) publishBlockDone(blockNumber *big.Int) {
	ch, err := s.mqConn.Channel()
	if err != nil {
		logger.GetLogger().Sugar().Errorf("Failed to open a channel: %v", err)
		return
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"blockNumber_done_queue",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logger.GetLogger().Sugar().Errorf("Failed to declare a queue: %v", err)
	}
	err = ch.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(blockNumber.String()),
		},
	)
	if err != nil {
		logger.GetLogger().Sugar().Errorf("Failed to publish a message: %v", err)
		return
	}
	logger.GetLogger().Sugar().Infof("Published a message: %s", blockNumber.String())

}

func (s *Service) scan(ctx context.Context) {
	for {
		select {
		case job := <-s.jobChan:
			blockNumber := job.BlockNumber
			block, err := s.blockScanner.BlockByNumber(ctx, blockNumber)

			if err != nil {
				logger.GetLogger().Sugar().Errorf("scan block %d error: %s", blockNumber, err.Error())
				job.DoneChan <- model.JobResult{
					BlockNumber: nil,
					Msg:         job.Msg,
				}
				continue
			}
			err = s.scanBlockInfo(ctx, block)
			if err != nil {
				logger.GetLogger().Sugar().Errorf("scan block %d error: %s", blockNumber, err.Error())
				job.DoneChan <- model.JobResult{
					BlockNumber: nil,
					Msg:         job.Msg,
				}
				continue
			}
			job.DoneChan <- model.JobResult{
				BlockNumber: blockNumber,
				Msg:         job.Msg,
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) scanBlockInfo(ctx context.Context, block *types.Block) error {
	blockRow := convertBlockToRow(block)
	// sned to chan

	s.blockConsumer.GetChan().(chan model.BlockRow) <- blockRow

	for _, tx := range block.Transactions() {
		txRow, err := convertTxToRow(tx, *block.Number())
		if err != nil {
			logger.LoadExtra(map[string]interface{}{
				"err": err.Error(),
			}).Error("convert tx to row error")
			return fmt.Errorf("scanBlockInfo: %s", err.Error())
		}

		s.txConsumer.GetChan().(chan model.TransactionRow) <- txRow

		logs, err := s.txScanner.LogsByTxHash(ctx, tx.Hash())
		if err != nil && !strings.Contains(err.Error(), "LogsByTxHash : not found") {
			logger.LoadExtra(map[string]interface{}{
				"err": err.Error(),
			}).Error("get logs error")
			return fmt.Errorf("scanBlockInfo: %s", err.Error())
		}

		for _, log := range logs {
			logRow := convertLogToRow(log, tx.Hash().Hex())
			s.logConsumer.GetChan().(chan model.LogRow) <- logRow
		}
	}
	return nil
}

func (s *Service) Shutdown() {
	s.shutdownOnce.Do(func() {

		s.mqConn.Close()
		s.cancel()
		close(s.jobChan)
		close(s.jobDone)
		s.blockConsumer.Shutdown()
		s.txConsumer.Shutdown()
		s.logConsumer.Shutdown()
		s.blockScanner.Shutdown()
		s.txScanner.Shutdown()

	})
}
