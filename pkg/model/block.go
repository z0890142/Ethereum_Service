package model

import (
	"math/big"

	"github.com/streadway/amqp"
)

type TransactionWithLogs struct {
	TxHash string `json:"tx_hash"`
	From   string `json:"from"`
	To     string `json:"to"`

	Nonce uint64 `json:"nonce"`
	Data  string `json:"data"`
	Value string `json:"value"`
	Logs  []Log  `json:"logs"`
}

type Log struct {
	Index int    `json:"index"`
	Data  string `json:"data"`
}

type Job struct {
	BlockNumber *big.Int
	DoneChan    chan JobResult
	Msg         *amqp.Delivery
}

type JobResult struct {
	BlockNumber *big.Int
	Msg         *amqp.Delivery
}
