package scanner

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TxScanner interface {
	LogsByTxHash(ctx context.Context, txHash common.Hash) ([]*types.Log, error)
}

type defaultTxScanner struct {
	ethClient *ethclient.Client
}

func NewDefaultTxScanner(ethClient *ethclient.Client) TxScanner {
	return &defaultTxScanner{
		ethClient: ethClient,
	}
}

func (s *defaultTxScanner) LogsByTxHash(ctx context.Context, txHash common.Hash) ([]*types.Log, error) {
	_, isPending, err := s.ethClient.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("LogsByTxHash : %w", err)
	}
	if isPending {
		return nil, nil
	}
	receipt, err := s.ethClient.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("LogsByTxHash : %w", err)
	}
	return receipt.Logs, nil
}
