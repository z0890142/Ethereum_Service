package model

type BlockRow struct {
	Hash       string
	Number     int64
	GasLimit   uint64
	GasUsed    uint64
	Difficulty int64
	Time       uint64
	Nonce      uint64
	Root       string
	ParentHash string
	TxHash     string
	UncleHash  string
	Extra      []byte
}

type TransactionRow struct {
	Hash        string
	BlockNumber int64
	Nonce       uint64
	To          string
	From        string
	Value       int64
	Data        []byte
}

type LogRow struct {
	TxHash string
	Index  uint
	Data   []byte
}

type LatestBlockNumber struct {
	Id          int64
	BlockNumber int64
}
