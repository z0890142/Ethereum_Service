package indexer_service

import (
	"Ethereum_Service/config"
	"Ethereum_Service/pkg/utils/logger"
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/streadway/amqp"
)

const (
	storeInterval = 10 * time.Second
)

type Service struct {
	ethClient *ethclient.Client

	scanHandler ScanHandler
	mqConn      *amqp.Connection
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
	scanHandler := NewScanHandler(rcpEndpoint)
	s := Service{
		ethClient:   ethClient,
		scanHandler: scanHandler,
		rcpEndpoint: rcpEndpoint,
	}

	return &s, nil
}

func (s *Service) Start(workerCount int) {
	logger.GetLogger().Sugar().Infof("start indexer service with %d workers", workerCount)
	s.shutDownCtx, s.cancel = context.WithCancel(context.Background())

	err := s.createMqConn()
	if err != nil {
		logger.GetLogger().Sugar().Fatalf("Failed to create RabbitMQ connection: %v", err)
		return
	}

	// s.startConsumers()
	s.startScanners(workerCount)
}

func (s *Service) createMqConn() error {
	conn, err := amqp.Dial(config.GetConfig().MQEndpoint)
	if err != nil {
		return err
	}

	s.mqConn = conn
	return nil
}

func (s *Service) startScanners(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go s.scanHandler.Scan(s.shutDownCtx, s.ethClient, s.mqConn)
	}
}

func (s *Service) publishBlockDone(blockNumber *big.Int, msg *amqp.Delivery) {
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
	msg.Ack(true)
	logger.GetLogger().Sugar().Infof("Published a message: %s", blockNumber.String())

}

func (s *Service) Shutdown() {
	s.shutdownOnce.Do(func() {
		s.mqConn.Close()
		s.scanHandler.Shutdown()
		s.cancel()
	})
}
