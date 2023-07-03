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
	TxDetailByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)
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

func (s *defaultTxScanner) TxDetailByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
	var isPending bool
	var err error
	var tx *types.Transaction
	for i := 0; i < config.GetConfig().MaxRetryTime; i++ {
		tx, isPending, err = s.ethClient.TransactionByHash(ctx, txHash)
		if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
			s.createClient()
		}
		if err == nil {
			break
		}
		<-time.NewTimer(time.Second * 1).C
	}

	if err != nil {
		return tx, isPending, fmt.Errorf("LogsByTxHash : %w", err)
	}
	return tx, isPending, nil
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
