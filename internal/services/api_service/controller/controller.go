package controller

import (
	"Ethereum_Service/config"
	"Ethereum_Service/internal/data"
	"Ethereum_Service/internal/scanner"
	"Ethereum_Service/pkg/model"
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	mysqlHandler data.DataHandler
	txScanner    scanner.TxScanner
	blockScanner scanner.BlockScanner

	ethClient *ethclient.Client
}

func NewController() *Controller {
	ethClient, err := ethclient.Dial(config.GetConfig().RCPEndpoint)
	if err != nil {
		panic(err)
	}

	mysqlHandler, err := data.NewMysqlHandler(&config.GetConfig().Databases)
	if err != nil {
		panic(err)
	}

	txScanner := scanner.NewDefaultTxScanner(config.GetConfig().RCPEndpoint)
	blockScanner := scanner.NewDefaultBlockScanner(config.GetConfig().RCPEndpoint)

	return &Controller{
		ethClient:    ethClient,
		mysqlHandler: mysqlHandler,
		txScanner:    txScanner,
		blockScanner: blockScanner,
	}
}

func (c *Controller) GetTransaction(ginC *gin.Context) {
	txHash := ginC.Param("txHash")
	resp, err := c.getTxFromDB(txHash)
	if err == nil {
		ginC.JSON(200, resp)
		return
	}

	resp, err = c.getTxFromRPC(txHash)
	if err != nil {
		ginC.JSON(404, gin.H{"error": err.Error()})
		return
	}
	ginC.JSON(200, resp)
}
func (c *Controller) ListBlocks(ginC *gin.Context) {
	limit := ginC.Query("limit")
	limitUint, err := strconv.ParseUint(limit, 10, 64)
	if err != nil {
		ginC.JSON(400, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.listBlocks(limitUint)
	if err != nil {
		ginC.JSON(404, gin.H{"error": err.Error()})
		return
	}
	ginC.JSON(200, resp)
}
func (c *Controller) GetBlock(ginC *gin.Context) {
	blockId := ginC.Param("id")

	blockNum, err := strconv.ParseInt(blockId, 10, 64)
	if err != nil {
		ginC.JSON(400, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.getBlockDetail(blockNum)
	if err != nil {
		ginC.JSON(404, gin.H{"error": err.Error()})
		return
	}
	ginC.JSON(200, resp)

}
func (c *Controller) Shutdown() {}

func (c *Controller) getTxFromDB(txHash string) (model.TxResponse, error) {
	txRow := model.TransactionRow{
		Hash: txHash,
	}
	err := c.mysqlHandler.GetTransactionRow(context.Background(), &txRow)
	if err != nil {
		return model.TxResponse{}, err
	}

	resp := model.TxResponse{
		TxHash: txRow.Hash,
		From:   txRow.From,
		To:     txRow.To,
		Value:  strconv.FormatInt(txRow.Value, 10),
		Data:   string(txRow.Data),
	}

	logs, err := c.mysqlHandler.GetLogRowByTxHash(context.Background(), txHash)
	resp.Logs = convertLogRowToResp(logs)
	return resp, err
}

func (c *Controller) getTxFromRPC(txHash string) (model.TxResponse, error) {
	hash := common.HexToHash(txHash)
	tx, logs, err := c.txScanner.TxDetailByHash(context.Background(), hash)
	if err != nil {
		return model.TxResponse{}, err
	}

	from, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err != nil {
		return model.TxResponse{}, fmt.Errorf("getTxFromRPC : %w", err)
	}
	resp := model.TxResponse{
		TxHash: tx.Hash().Hex(),
		From:   from.Hex(),
		To:     tx.To().Hex(),
		Value:  tx.Value().String(),
		Data:   string(tx.Data()),
	}
	resp.Logs = convertTypeLogToResp(logs)

	receipt, err := c.ethClient.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return model.TxResponse{}, fmt.Errorf("getTxFromRPC : %w", err)
	}

	txRow := model.TransactionRow{
		Hash:        tx.Hash().Hex(),
		From:        from.Hex(),
		To:          tx.To().Hex(),
		Value:       tx.Value().Int64(),
		Data:        tx.Data(),
		Nonce:       tx.Nonce(),
		BlockNumber: receipt.BlockNumber.Int64(),
	}
	go c.mysqlHandler.SaveTransactionRow(context.Background(), []*model.TransactionRow{&txRow})
	return resp, nil
}

func convertLogRowToResp(logRows []model.LogRow) []model.LogResponse {
	resp := make([]model.LogResponse, 0, len(logRows))
	for _, logRow := range logRows {
		resp = append(resp, model.LogResponse{
			Index: logRow.Index,
			Data:  string(logRow.Data),
		})
	}
	return resp
}

func convertTypeLogToResp(logs []*types.Log) []model.LogResponse {
	resp := make([]model.LogResponse, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, model.LogResponse{
			Index: log.Index,
			Data:  string(log.Data),
		})
	}
	return resp
}

func (c *Controller) getBlockDetail(blockNumber int64) (model.BlockResponseWithTx, error) {
	resp, err := c.getBlockFromDB(blockNumber)
	if err == nil {
		return resp, err
	}
	resp, err = c.getBlockFromRPC(blockNumber)
	if err != nil {
		return model.BlockResponseWithTx{}, err
	}
	return resp, nil
}

func (c *Controller) getBlockFromDB(blockNumber int64) (model.BlockResponseWithTx, error) {
	blockRow := model.BlockRow{
		Number: blockNumber,
	}
	err := c.mysqlHandler.GetBlockRow(context.Background(), &blockRow)
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

	txRows, err := c.mysqlHandler.GetTransactionRowByBlockNumber(context.Background(), blockNumber)
	for _, tx := range txRows {
		resp.Transactions = append(resp.Transactions, tx.Hash)
	}
	return resp, err
}

func (c *Controller) getBlockFromRPC(blockNumber int64) (model.BlockResponseWithTx, error) {
	blockNumBig := big.NewInt(blockNumber)
	block, err := c.blockScanner.BlockByNumber(context.Background(), blockNumBig)
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
	go c.mysqlHandler.SaveBlockRows(context.Background(), []*model.BlockRow{&blockRow})
	return resp, nil
}

func (c *Controller) listBlocks(limit uint64) ([]model.BlockResponse, error) {
	latestBlockNumber, err := c.ethClient.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}
	resp, err := c.listBlocksFromDB(latestBlockNumber, limit)
	if len(resp) != 0 && err == nil {
		return resp, err
	}
	resp, err = c.listBlocksFromRPC(latestBlockNumber, limit)
	return resp, nil
}

func (c *Controller) listBlocksFromDB(latestBlockNumber, limit uint64) ([]model.BlockResponse, error) {
	numbers := make([]int64, 0, limit)
	for i := latestBlockNumber; i > latestBlockNumber-limit; i-- {
		numbers = append(numbers, int64(i))
	}
	blockRows, err := c.mysqlHandler.GetBlockRowByBlockNumbers(context.Background(), numbers)
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

func (c *Controller) listBlocksFromRPC(latestBlockNumber, limit uint64) ([]model.BlockResponse, error) {
	resp := make([]model.BlockResponse, 0, limit)
	blockRows := make([]*model.BlockRow, 0)
	for i := latestBlockNumber; i > latestBlockNumber-limit; i-- {
		number := big.NewInt(int64(i))
		block, err := c.blockScanner.BlockByNumber(context.Background(), number)
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
	go c.mysqlHandler.SaveBlockRows(context.Background(), blockRows)

	return resp, nil
}
