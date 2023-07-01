package scanner

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockScanner interface {
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}

type DefaultBlockScanner struct {
	ethClient *ethclient.Client
}

func NewDefaultBlockScanner(ethClient *ethclient.Client) BlockScanner {
	return &DefaultBlockScanner{
		ethClient: ethClient,
	}
}

func (s *DefaultBlockScanner) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	b, err := s.ethClient.BlockByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("BlockByNumber : %w", err)
	}
	return b, nil
}

func (s *DefaultBlockScanner) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	b, err := s.ethClient.BlockByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("BlockByHash : %w", err)
	}
	return b, nil
}
