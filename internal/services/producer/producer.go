package producer

import (
	"Ethereum_Service/config"
	"Ethereum_Service/internal/data"
	"Ethereum_Service/pkg/utils/logger"
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/streadway/amqp"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Producer struct {
	ethClient           *ethclient.Client
	mysqlHandler        data.DataHandler
	mqConn              *amqp.Connection
	latestBlockNumber   uint64
	dbLatestBlockNumber int64
}

const (
	BlockNumberQueueName       = "blockNumber_queue"
	BlockNumberDoneQueueName   = "blockNumber_done_queue"
	ProducerServiceConsumerTag = "producer_service"
)

var (
	ErrMaxRetryExceeded = errors.New("max retry attempts exceeded")
)

func NewProducer(rcpEndpoint string) (*Producer, error) {
	ethClient, err := ethclient.Dial(rcpEndpoint)
	if err != nil {
		return nil, fmt.Errorf("NewProducer : %s", err.Error())
	}

	mysqlHandler, err := data.NewMysqlHandler(&config.GetConfig().Databases)
	if err != nil {
		return nil, fmt.Errorf("NewProducer : %s", err.Error())
	}

	return &Producer{
		ethClient:    ethClient,
		mysqlHandler: mysqlHandler,
	}, nil
}

func (p *Producer) Start() {
	p.createEthClient()
	p.getLatestBlockNumber()
	go p.continueUpdateBlockNumber()
	go p.receiveACK()
	p.startLoop()
}

func (p *Producer) startLoop() {
	index, err := p.mysqlHandler.GetLatestBlockNumber(context.Background())
	p.dbLatestBlockNumber = index
	if err != nil {
		logger.GetLogger().Sugar().Errorf("startLoop: failed to get latest block number from MySQL: %s", err.Error())
		return
	}

	for index <= int64(p.latestBlockNumber) {
		blockNumber := strconv.FormatInt(index, 10)
		err := p.pushMsg(blockNumber)
		if err != nil {
			logger.GetLogger().Sugar().Errorf("Start : %s", err.Error())
		}
		index++
	}
}

func (p *Producer) createEthClient() {
	var err error
	p.mqConn, err = amqp.Dial(config.GetConfig().MQEndpoint)
	if err != nil {
		panic(err)
	}
}

func (p *Producer) getLatestBlockNumber() {

	var err error
	for i := 0; i < config.GetConfig().MaxRetryTime; i++ {
		p.latestBlockNumber, err = p.ethClient.BlockNumber(context.Background())
		if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
			p.createEthClient()
		}
		if err == nil {
			break
		}
		<-time.NewTimer(time.Second * 1).C
	}
	if err != nil {
		panic(err)
	}

}
func (p *Producer) continueUpdateBlockNumber() {
	t := time.NewTicker(time.Second * 5)
	for {
		var err error
		for i := 0; i < config.GetConfig().MaxRetryTime; i++ {
			p.latestBlockNumber, err = p.ethClient.BlockNumber(context.Background())
			if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
				p.createEthClient()
			}
			if err == nil {
				break
			}
			<-time.NewTimer(time.Second * 1).C
		}
		if err != nil {
			panic(err)
		}
		<-t.C
	}
}

func (p *Producer) pushMsg(blockNumber string) error {
	ch, err := p.mqConn.Channel()
	if err != nil {
		return fmt.Errorf("Start : %s", err.Error())
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		BlockNumberQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("pushMsg: failed to declare queue: %s", err.Error())
	}

	err = ch.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(blockNumber),
		},
	)
	if err != nil {
		return fmt.Errorf("pushMsg: failed to publish message: %s", err.Error())
	}
	return nil
}

func (p *Producer) receiveACK() {

	var msgs <-chan amqp.Delivery
	for {
		ch, err := p.mqConn.Channel()
		if err != nil {
			log.Fatalf("Failed to open a channel: %v", err)
		}
		msgs, err = ch.Consume(
			"blockNumber_done_queue",
			"producer_service",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil && strings.Contains(err.Error(), "no queue") {
			continue
		}
		if err == nil {
			break
		}
	}

	for msg := range msgs {
		logger.GetLogger().Sugar().Infof("receive : %s", msg.Body)
		num, err := strconv.ParseInt(string(msg.Body), 10, 64)
		if err != nil {
			logger.GetLogger().Sugar().Errorf("receiveACK : %s", err.Error())
			continue
		}

		if num < p.dbLatestBlockNumber {
			msg.Ack(true)
			continue
		}

		err = p.mysqlHandler.UpdateLatestBlockNumber(context.Background(), num)
		if err != nil {
			logger.GetLogger().Sugar().Errorf("receiveACK : %s", err.Error())
		}

		p.dbLatestBlockNumber = num
		msg.Ack(true)
		logger.GetLogger().Sugar().Infof("receiveACK : %d", num)
	}
}
