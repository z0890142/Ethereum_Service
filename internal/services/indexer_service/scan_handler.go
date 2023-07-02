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

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/streadway/amqp"
)

type ScanHandler struct {
	blockScanner scanner.BlockScanner
	txScanner    scanner.TxScanner
	logScanner   scanner.LogScanner

	blockConsumer consumer.Consumer
	txConsumer    consumer.Consumer
	logConsumer   consumer.Consumer
}

func NewScanHandler(rcpEndpoint string) ScanHandler {

	blockScanner := scanner.NewDefaultBlockScanner(rcpEndpoint)
	txScanner := scanner.NewDefaultTxScanner(rcpEndpoint)
	logScanner := scanner.NewDefaultLogScanner(rcpEndpoint)

	blockConsume := consumer.NewConsumer(&consumer.ConsumerConf{
		Type:            c.BlockConsumerType,
		StoreBufferSize: config.GetConfig().StoreBufferSize,
		StoreInterval:   storeInterval,
	})

	txConsume := consumer.NewConsumer(&consumer.ConsumerConf{
		Type:            c.TxConsumerType,
		StoreBufferSize: config.GetConfig().StoreBufferSize,
		StoreInterval:   storeInterval,
	})
	logConsume := consumer.NewConsumer(&consumer.ConsumerConf{
		Type:            c.LogConsumerType,
		StoreBufferSize: config.GetConfig().StoreBufferSize,
		StoreInterval:   storeInterval,
	})

	go blockConsume.Run()
	go txConsume.Run()
	go logConsume.Run()

	return ScanHandler{

		blockScanner: blockScanner,
		txScanner:    txScanner,
		logScanner:   logScanner,

		blockConsumer: blockConsume,
		txConsumer:    txConsume,
		logConsumer:   logConsume,
	}
}
func (s *ScanHandler) Scan(ctx context.Context, ethClient *ethclient.Client, mqConn *amqp.Connection) {
	ch, err := mqConn.Channel()
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
		blockNumberBig := big.NewInt(int64(blockNumber))
		block, err := ethClient.BlockByNumber(ctx, blockNumberBig)

		if err != nil {
			logger.GetLogger().Sugar().Errorf("scan block %d error: %s", blockNumber, err.Error())
			msg.Ack(false)
			continue
		}

		err = s.getBlockInfo(ctx, block)
		if err != nil {
			logger.GetLogger().Sugar().Errorf("scan block %d error: %s", blockNumber, err.Error())
			msg.Ack(false)
			continue
		}

		s.scanDone(mqConn, blockNumberBig, &msg)
	}
}

func (s *ScanHandler) getBlockInfo(ctx context.Context, block *types.Block) error {

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

		_, isPending, err := s.txScanner.TxDetailByHash(ctx, tx.Hash())
		if err != nil && !strings.Contains(err.Error(), "LogsByTxHash : not found") {
			logger.LoadExtra(map[string]interface{}{
				"err": err.Error(),
			}).Error("get logs error")
			return fmt.Errorf("scanBlockInfo: %s", err.Error())
		}
		if isPending {
			continue
		}
		logs, err := s.logScanner.GetLogs(ctx, tx.Hash())
		if err != nil {
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

func (s *ScanHandler) scanDone(conn *amqp.Connection, blockNumber *big.Int, msg *amqp.Delivery) {
	ch, err := conn.Channel()
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
		return
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

func (s *ScanHandler) Shutdown() {
	s.blockConsumer.Shutdown()
	s.txConsumer.Shutdown()
	s.logConsumer.Shutdown()
	s.blockScanner.Shutdown()
	s.txScanner.Shutdown()
}
