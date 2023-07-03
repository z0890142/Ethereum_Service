package controller

import (
	"Ethereum_Service/internal/data"
	"Ethereum_Service/internal/scanner"
	"Ethereum_Service/pkg/model"

	"context"
	"math/big"
)

func listBlocksFromStore(dataHandler data.DataHandler, latestBlockNumber, limit uint64) ([]model.BlockResponse, error) {
	numbers := make([]int64, 0, limit)
	for i := latestBlockNumber; i > latestBlockNumber-limit; i-- {
		numbers = append(numbers, int64(i))
	}
	blockRows, err := dataHandler.GetBlockRowByBlockNumbers(context.Background(), numbers)
	if err != nil {
		return nil, err
	}

	resp := make([]model.BlockResponse, 0, len(blockRows))
	for _, blockRow := range blockRows {
		resp = append(resp, model.BlockResponse{
			BlockNum:   blockRow.Number,
			BlockHash:  blockRow.Hash,
			ParentHash: blockRow.ParentHash,
			BlockTime:  blockRow.Time,
		})
	}
	return resp, nil
}

func listBlocksFromRPC(mysqlHandler, redesHandler data.DataHandler, blockScanner scanner.BlockScanner, latestBlockNumber, limit uint64) ([]model.BlockResponse, error) {
	resp := make([]model.BlockResponse, 0, limit)
	blockRows := make([]*model.BlockRow, 0)
	for i := latestBlockNumber; i > latestBlockNumber-limit; i-- {
		number := big.NewInt(int64(i))
		block, err := blockScanner.BlockByNumber(context.Background(), number)
		if err != nil {
			return nil, err
		}
		resp = append(resp, model.BlockResponse{
			BlockNum:   block.Number().Int64(),
			BlockHash:  block.Hash().Hex(),
			ParentHash: block.ParentHash().Hex(),
			BlockTime:  block.Time(),
		})
		blockRows = append(blockRows, &model.BlockRow{
			Hash:       block.Hash().Hex(),
			Number:     (*block.Number()).Int64(),
			GasLimit:   block.GasLimit(),
			GasUsed:    block.GasUsed(),
			Difficulty: (*block.Difficulty()).Int64(),
			Time:       block.Time(),
			Nonce:      block.Nonce(),
			Root:       block.Root().Hex(),
			ParentHash: block.ParentHash().Hex(),
			TxHash:     block.TxHash().Hex(),
			UncleHash:  block.UncleHash().Hex(),
			Extra:      block.Extra(),
		})
	}
	go mysqlHandler.SaveBlockRows(context.Background(), blockRows)
	go redesHandler.SaveBlockRows(context.Background(), blockRows)

	return resp, nil
}
