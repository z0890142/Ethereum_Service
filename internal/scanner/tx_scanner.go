package scanner

import (
	"Ethereum_Service/config"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TxScanner interface {
	LogsByTxHash(ctx context.Context, txHash common.Hash) ([]*types.Log, error)
	Shutdown()
}

type defaultTxScanner struct {
	ethClient   *ethclient.Client
	rpcEndpoint string
}

func NewDefaultTxScanner(rpcEndpoint string) TxScanner {
	s := &defaultTxScanner{
		rpcEndpoint: rpcEndpoint,
	}
	s.createClient()
	return s
}

func (s *defaultTxScanner) LogsByTxHash(ctx context.Context, txHash common.Hash) ([]*types.Log, error) {
	var isPending bool
	var err error
	for i := 0; i < config.GetConfig().MaxRetryTime; i++ {
		_, isPending, err = s.ethClient.TransactionByHash(ctx, txHash)
		if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
			s.createClient()
		}
		if err == nil {
			break
		}
		<-time.NewTimer(time.Second * 1).C
	}

	if err != nil {
		return nil, fmt.Errorf("LogsByTxHash : %w", err)
	}
	if isPending {
		return nil, nil
	}

	var receipt *types.Receipt
	for i := 0; i < config.GetConfig().MaxRetryTime; i++ {
		receipt, err = s.ethClient.TransactionReceipt(ctx, txHash)
		if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
			s.createClient()
		}
		if err == nil {
			break
		}
		<-time.NewTimer(time.Second * 1).C
	}

	if err != nil {
		return nil, fmt.Errorf("LogsByTxHash : %w", err)
	}
	return receipt.Logs, nil
}

func (s *defaultTxScanner) createClient() {
	var err error
	s.ethClient, err = ethclient.Dial(s.rpcEndpoint)
	if err != nil {
		panic(err)
	}
}

func (s *defaultTxScanner) Shutdown() {
	s.ethClient.Close()
}
