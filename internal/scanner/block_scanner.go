package scanner

import (
	"Ethereum_Service/config"
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockScanner interface {
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	Shutdown()
}

type DefaultBlockScanner struct {
	ethClient   *ethclient.Client
	rpcEndpoint string
}

func NewDefaultBlockScanner(rpcEndpoint string) BlockScanner {
	s := &DefaultBlockScanner{
		rpcEndpoint: rpcEndpoint,
	}
	s.createClient()
	return s
}

func (s *DefaultBlockScanner) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	var b *types.Block
	var err error
	for i := 0; i < config.GetConfig().MaxRetryTime; i++ {
		b, err = s.ethClient.BlockByNumber(ctx, number)
		if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
			s.createClient()
		}
		if err == nil {
			break
		}
		<-time.NewTimer(time.Second * 1).C
	}
	if err != nil {
		return nil, fmt.Errorf("BlockByNumber : %w", err)
	}
	return b, nil
}

func (s *DefaultBlockScanner) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	var b *types.Block
	var err error
	for i := 0; i < config.GetConfig().MaxRetryTime; i++ {
		b, err = s.ethClient.BlockByHash(ctx, hash)
		if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
			s.createClient()
		}
		if err == nil {
			break
		}
		<-time.NewTimer(time.Second * 1).C
	}
	if err != nil {
		return nil, fmt.Errorf("BlockByHash : %w", err)
	}
	return b, nil
}

func (s *DefaultBlockScanner) createClient() {
	var err error
	s.ethClient, err = ethclient.Dial(s.rpcEndpoint)
	if err != nil {
		panic(err)
	}
}

func (s *DefaultBlockScanner) Shutdown() {
	s.ethClient.Close()
}
