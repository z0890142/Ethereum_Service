package model

type TxResponse struct {
	TxHash string        `json:"tx_hash"`
	From   string        `json:"from"`
	To     string        `json:"to"`
	Data   string        `json:"data"`
	Value  string        `json:"value"`
	Nonce  uint64        `json:"nonce"`
	Logs   []LogResponse `json:"logs"`
}

type LogResponse struct {
	Index uint   `json:"index"`
	Data  string `json:"data"`
}

type BlockResponseWithTx struct {
	BlockNum     int64    `json:"block_num"`
	BlockHash    string   `json:"block_hash"`
	BlockTime    uint64   `json:"block_time"`
	ParentHash   string   `json:"parent_hash"`
	Transactions []string `json:"transactions"`
}

type BlockResponse struct {
	BlockNum   int64  `json:"block_num"`
	BlockHash  string `json:"block_hash"`
	BlockTime  uint64 `json:"block_time"`
	ParentHash string `json:"parent_hash"`
}
