package controller

import (
	"Ethereum_Service/internal/data"
	"Ethereum_Service/internal/scanner"
	"Ethereum_Service/pkg/model"
	"context"
	"math/big"
)

func getBlockFromStore(dataHandler data.DataHandler, blockNumber int64) (model.BlockResponseWithTx, error) {
	blockRow := model.BlockRow{
		Number: blockNumber,
	}
	err := dataHandler.GetBlockRow(context.Background(), &blockRow)
	if err != nil {
		return model.BlockResponseWithTx{}, err
	}

	resp := model.BlockResponseWithTx{
		BlockNum:     blockNumber,
		BlockHash:    blockRow.Hash,
		ParentHash:   blockRow.ParentHash,
		BlockTime:    blockRow.Time,
		Transactions: []string{},
	}

	txRows, err := dataHandler.GetTransactionRowByBlockNumber(context.Background(), blockNumber)
	for _, tx := range txRows {
		resp.Transactions = append(resp.Transactions, tx.Hash)
	}
	return resp, err
}

func getBlockFromRPC(blockScanner scanner.BlockScanner, mysqlHandler, redisHandler data.DataHandler, blockNumber int64) (model.BlockResponseWithTx, error) {
	blockNumBig := big.NewInt(blockNumber)
	block, err := blockScanner.BlockByNumber(context.Background(), blockNumBig)
	if err != nil {
		return model.BlockResponseWithTx{}, err
	}
	resp := model.BlockResponseWithTx{
		BlockNum:     blockNumber,
		BlockHash:    block.Hash().Hex(),
		ParentHash:   block.ParentHash().Hex(),
		BlockTime:    block.Time(),
		Transactions: []string{},
	}

	txs := block.Transactions()
	for _, tx := range txs {
		resp.Transactions = append(resp.Transactions, tx.Hash().Hex())
	}
	blockRow := model.BlockRow{
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
	}
	go mysqlHandler.SaveBlockRows(context.Background(), []*model.BlockRow{&blockRow})
	go redisHandler.SaveBlockRows(context.Background(), []*model.BlockRow{&blockRow})
	return resp, nil
}
