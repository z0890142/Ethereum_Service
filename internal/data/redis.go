package data

import (
	"Ethereum_Service/config"
	"Ethereum_Service/pkg/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
)

type RedisDataHandler struct {
	redisClient *redis.Client
}

const (
	TransactionRowKey = "Transaction"
)

func NewRedisDataHandler() *RedisDataHandler {
	addr := fmt.Sprintf("%s:%s", config.GetConfig().Redis.Host, config.GetConfig().Redis.Port)

	cli := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.GetConfig().Redis.Password,
	})
	return &RedisDataHandler{redisClient: cli}
}

func (h *RedisDataHandler) GetTransactionRow(ctx context.Context, tx *model.TransactionRow) error {
	key := fmt.Sprintf("%d:TransactionRow:%s", tx.BlockNumber, tx.Hash)
	txString, err := h.redisClient.Get(key).Result()
	if err != nil {
		return fmt.Errorf("GetTransactionRow: %w", err)
	}
	bs := []byte(txString)
	err = json.Unmarshal(bs, tx)
	if err != nil {
		return fmt.Errorf("GetTransactionRow: %w", err)
	}
	return nil
}

func (h *RedisDataHandler) SaveTransactionRow(ctx context.Context, txRow []*model.TransactionRow) error {
	for _, tx := range txRow {
		key := fmt.Sprintf("%d:TransactionRow:%s", tx.BlockNumber, tx.Hash)
		bs, err := json.Marshal(*tx)
		if err != nil {
			return fmt.Errorf("SaveTransactionRow: %w", err)
		}
		err = h.redisClient.Set(key, string(bs), 0).Err()
		if err != nil {
			return fmt.Errorf("SaveTransactionRow: %w", err)
		}
	}
	return nil
}

func (h *RedisDataHandler) GetBlockRowByBlockNumbers(ctx context.Context, numbers []int64) ([]model.BlockRow, error) {
	var result []model.BlockRow
	for _, number := range numbers {
		key := fmt.Sprintf("BlockRow:%d", number)
		v, err := h.redisClient.Get(key).Result()
		if err != nil {
			return nil, fmt.Errorf("GetBlockRowByBlockNumbers: %w", err)
		}
		bs := []byte(v)

		var blockRow model.BlockRow
		err = json.Unmarshal(bs, &blockRow)
		if err != nil {
			return nil, fmt.Errorf("GetBlockRowByBlockNumbers: %w", err)
		}
		result = append(result, blockRow)
	}
	return result, nil
}

func (h *RedisDataHandler) GetBlockRow(ctx context.Context, block *model.BlockRow) error {
	key := fmt.Sprintf("BlockRow:%d", block.Number)

	blockString, err := h.redisClient.Get(key).Result()
	if err != nil {
		return fmt.Errorf("GetBlockRow: %w", err)
	}
	bs := []byte(blockString)

	err = json.Unmarshal(bs, block)
	if err != nil {
		return fmt.Errorf("GetBlockRow: %w", err)
	}
	return nil
}

func (h *RedisDataHandler) SaveBlockRows(ctx context.Context, blocks []*model.BlockRow) error {
	for _, block := range blocks {
		key := fmt.Sprintf("BlockRow:%d", block.Number)
		bs, err := json.Marshal(*block)
		if err != nil {
			return fmt.Errorf("SaveBlockRows: %w", err)
		}
		err = h.redisClient.Set(key, string(bs), 0).Err()
		if err != nil {
			return fmt.Errorf("SaveBlockRows: %w", err)
		}
	}
	return nil
}

func (h *RedisDataHandler) GetLogRowByTxHash(ctx context.Context, txHash string) ([]model.LogRow, error) {
	var result []model.LogRow
	key := fmt.Sprintf("TxLog:%s:*", txHash)
	iter := h.redisClient.Scan(0, key, 0).Iterator()
	for iter.Next() {
		key := iter.Val()
		v, err := h.redisClient.Get(key).Result()
		if err != nil {
			return nil, fmt.Errorf("GetLogRowByTxHash: %w", err)
		}
		bs := []byte(v)

		var logRow model.LogRow
		err = json.Unmarshal(bs, &logRow)
		if err != nil {
			return nil, fmt.Errorf("GetLogRowByTxHash: %w", err)
		}
		result = append(result, logRow)
	}
	return result, nil
}
func (h *RedisDataHandler) SaveLogRow(ctx context.Context, logRow []*model.LogRow) error {
	for _, log := range logRow {
		key := fmt.Sprintf("TxLog:%s:%s", log.TxHash, log.Index)
		bs, err := json.Marshal(log)
		if err != nil {
			return fmt.Errorf("SaveLogRow: %w", err)
		}
		err = h.redisClient.Set(key, string(bs), 0).Err()
		if err != nil {
			return fmt.Errorf("SaveLogRow: %w", err)
		}
	}
	return nil
}

func (h *RedisDataHandler) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	return 0, nil
}
func (h *RedisDataHandler) GetTransactionRowByBlockNumber(ctx context.Context, blockNumber int64) ([]model.TransactionRow, error) {
	return nil, nil
}

func (h *RedisDataHandler) UpdateLatestBlockNumber(ctx context.Context, blockNumber int64) error {
	return nil
}
