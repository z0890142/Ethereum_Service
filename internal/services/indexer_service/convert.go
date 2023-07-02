package indexer_service

import (
	"Ethereum_Service/pkg/model"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

func convertBlockToRow(block *types.Block) model.BlockRow {
	return model.BlockRow{
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
}

func convertTxToRow(tx *types.Transaction, blockNumber big.Int) (model.TransactionRow, error) {
	from, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err != nil {
		return model.TransactionRow{}, fmt.Errorf("convertTxToRow : %w", err)
	}
	txRow := model.TransactionRow{
		Hash:        tx.Hash().Hex(),
		BlockNumber: blockNumber.Int64(),
		Nonce:       tx.Nonce(),
		From:        from.Hex(),
		Value:       (*tx.Value()).Int64(),
		Data:        tx.Data(),
	}
	if tx.To() != nil {
		txRow.To = tx.To().Hex()
	}
	return txRow, nil
}

func convertLogToRow(log *types.Log, txHash string) model.LogRow {
	return model.LogRow{
		TxHash: txHash,
		Index:  log.Index,
		Data:   log.Data,
	}
}
