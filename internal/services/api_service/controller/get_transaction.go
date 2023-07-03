package controller

import (
	"Ethereum_Service/internal/data"
	"Ethereum_Service/internal/scanner"
	"Ethereum_Service/pkg/model"
	"context"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func getTxFromStore(dataHandler data.DataHandler, txHash string) (model.TxResponse, error) {
	txRow := model.TransactionRow{
		Hash: txHash,
	}
	err := dataHandler.GetTransactionRow(context.Background(), &txRow)
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

	logs, err := dataHandler.GetLogRowByTxHash(context.Background(), txHash)
	resp.Logs = convertLogRowToResp(logs)
	return resp, err
}

func getTxLogs(dataHandler data.DataHandler, txHash string) ([]model.LogRow, error) {
	logs, err := dataHandler.GetLogRowByTxHash(context.Background(), txHash)

	return logs, err
}
func getTxFromRPC(txScanner scanner.TxScanner, logScanner scanner.LogScanner, ethClient *ethclient.Client, mysqlHandler data.DataHandler, redisHandler data.DataHandler, txHash string) (model.TxResponse, error) {
	hash := common.HexToHash(txHash)
	tx, isPending, err := txScanner.TxDetailByHash(context.Background(), hash)
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

	receipt, err := ethClient.TransactionReceipt(context.Background(), hash)
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
	go mysqlHandler.SaveTransactionRow(context.Background(), []*model.TransactionRow{&txRow})
	go redisHandler.SaveTransactionRow(context.Background(), []*model.TransactionRow{&txRow})

	if !isPending {
		logs, err := logScanner.GetLogs(context.Background(), hash)
		if err != nil {
			return model.TxResponse{}, fmt.Errorf("getTxFromRPC : %w", err)
		}
		resp.Logs = convertTypeLogToResp(logs)
		logRows := convertTypeLogToRow(logs)
		go mysqlHandler.SaveLogRow(context.Background(), logRows)
		go redisHandler.SaveLogRow(context.Background(), logRows)
	}

	return resp, nil
}
