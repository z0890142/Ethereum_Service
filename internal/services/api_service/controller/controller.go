package controller

import (
	"Ethereum_Service/config"
	"Ethereum_Service/internal/data"
	"Ethereum_Service/internal/scanner"
	"Ethereum_Service/pkg/model"
	"context"
	"strconv"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	mysqlHandler data.DataHandler
	redisHandler data.DataHandler
	txScanner    scanner.TxScanner
	blockScanner scanner.BlockScanner
	logScanner   scanner.LogScanner

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

	redisHandler := data.NewRedisDataHandler()

	txScanner := scanner.NewDefaultTxScanner(config.GetConfig().RCPEndpoint)
	blockScanner := scanner.NewDefaultBlockScanner(config.GetConfig().RCPEndpoint)
	logScanner := scanner.NewDefaultLogScanner(config.GetConfig().RCPEndpoint)

	return &Controller{
		ethClient:    ethClient,
		mysqlHandler: mysqlHandler,
		redisHandler: redisHandler,
		txScanner:    txScanner,
		logScanner:   logScanner,
		blockScanner: blockScanner,
	}
}

func (c *Controller) GetTransaction(ginC *gin.Context) {
	txHash := ginC.Param("txHash")
	resp, err := c.getTransaction(txHash)
	if err != nil {
		ginC.JSON(404, gin.H{"error": err.Error()})
		return
	}
	logs, err := getTxLogs(c.redisHandler, txHash)
	if err != nil {
		ginC.JSON(404, gin.H{"error": err.Error()})
		return
	}
	resp.Logs = convertLogRowToResp(logs)
	ginC.JSON(200, resp)
}

func (c *Controller) getTransaction(txHash string) (model.TxResponse, error) {
	resp, err := getTxFromStore(c.redisHandler, txHash)
	if err == nil && resp.TxHash != "" {
		return resp, err
	}
	resp, err = getTxFromRPC(c.txScanner, c.logScanner, c.ethClient, c.mysqlHandler, c.redisHandler, txHash)
	return resp, err

}

func (c *Controller) getTxLogs(txHash string) ([]model.LogRow, error) {
	logRows, err := getTxLogs(c.redisHandler, txHash)
	if err == nil && len(logRows) != 0 {
		return logRows, err
	}
	logRows, err = getTxLogs(c.mysqlHandler, txHash)
	if err != nil {
		return nil, err
	}
	return logRows, nil
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

	txHashs, err := c.getBlockTx(blockNum)
	if err != nil {
		ginC.JSON(500, gin.H{"error": err.Error()})
		return
	}
	resp.Transactions = txHashs
	ginC.JSON(200, resp)

}
func (c *Controller) Shutdown() {}

func (c *Controller) getBlockDetail(blockNumber int64) (model.BlockResponseWithTx, error) {
	resp, err := getBlockFromStore(c.redisHandler, blockNumber)
	if err == nil && resp.BlockHash != "" {
		return resp, err
	}

	resp, err = getBlockFromStore(c.mysqlHandler, blockNumber)
	if err == nil && resp.BlockHash != "" {
		return resp, err
	}
	resp, err = getBlockFromRPC(c.blockScanner, c.mysqlHandler, c.redisHandler, blockNumber)
	if err != nil {
		return model.BlockResponseWithTx{}, err
	}
	return resp, nil
}
func (c *Controller) getBlockTx(blockNumber int64) ([]string, error) {
	txHashs, err := getTxHashFromStore(c.redisHandler, blockNumber)
	if err == nil && len(txHashs) != 0 {
		return txHashs, err
	}
	txHashs, err = getTxHashFromStore(c.mysqlHandler, blockNumber)
	if err == nil && len(txHashs) != 0 {
		return txHashs, err
	}
	txHashs, err = getTxHashFromRPC(c.blockScanner, blockNumber)
	if err != nil {
		return nil, err
	}
	return txHashs, nil

}

func (c *Controller) listBlocks(limit uint64) ([]model.BlockResponse, error) {
	latestBlockNumber, err := c.ethClient.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}
	resp, err := listBlocksFromStore(c.redisHandler, latestBlockNumber, limit)
	if len(resp) == int(limit) && err == nil {
		return resp, err
	}

	resp, err = listBlocksFromStore(c.mysqlHandler, latestBlockNumber, limit)
	if len(resp) == int(limit) && err == nil {
		return resp, err
	}
	resp, err = listBlocksFromRPC(c.mysqlHandler, c.redisHandler, c.blockScanner, latestBlockNumber, limit)
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

func convertTypeLogToRow(logs []*types.Log) []*model.LogRow {
	resp := make([]*model.LogRow, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, &model.LogRow{
			TxHash: log.TxHash.Hex(),
			Index:  log.Index,
			Data:   log.Data,
		})
	}
	return resp
}
