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

type LogScanner interface {
	GetLogs(ctx context.Context, txHash common.Hash) ([]*types.Log, error)
	Shutdown()
}

type defaultLogScanner struct {
	ethClient   *ethclient.Client
	rpcEndpoint string
}

func NewDefaultLogScanner(rpcEndpoint string) LogScanner {
	s := &defaultLogScanner{
		rpcEndpoint: rpcEndpoint,
	}
	s.createClient()
	return s
}

func (s *defaultLogScanner) createClient() {
	var err error
	s.ethClient, err = ethclient.Dial(s.rpcEndpoint)
	if err != nil {
		panic(err)
	}
}

func (s *defaultLogScanner) GetLogs(ctx context.Context, txHash common.Hash) ([]*types.Log, error) {
	var receipt *types.Receipt
	var err error
	for i := 0; i < config.GetConfig().MaxRetryTime; i++ {
		receipt, err = s.ethClient.TransactionReceipt(ctx, txHash)
		if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
			s.createClient()
		}
		if err == nil {
			break
		}
		<-time.NewTimer(time.Millisecond * 100).C
	}

	if err != nil {
		return nil, fmt.Errorf("LogsByTxHash : %w", err)
	}
	return receipt.Logs, nil
}

func (s *defaultLogScanner) Shutdown() {
	s.ethClient.Close()
}
