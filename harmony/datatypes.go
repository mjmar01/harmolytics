package harmony

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"math/big"
)

//<editor-fold desc="Basic types">

// Address contains wallet or contract address in one1... format as string
// and the same address as an eth address
type Address struct {
	OneAddress string
	EthAddress ethCommon.Address
}

// Token contains token information as defined by the HRC-20 standard
type Token struct {
	Address  Address
	Name     string
	Symbol   string
	Decimals int
}

// Method contains information of a smart contract method and its parameters
type Method struct {
	Signature  string
	Name       string
	Parameters []string
}

//</editor-fold>

//<editor-fold desc="Transaction related types">

const (
	TxSuccessful = 1
	TxFailed     = 0
)

// TransactionLog contains the raw data of an event in hex format,
// the transaction hash and log index which uniquely identify the log
// and the associated Address
type TransactionLog struct {
	TxHash   string
	LogIndex int
	Address  Address
	Topics   []string
	Data     string
}

// Transaction contains all relevant information of a transaction
type Transaction struct {
	TxHash    string
	Sender    Address
	Receiver  Address
	BlockNum  uint64
	Timestamp uint64
	Value     *big.Int
	Method    Method
	Input     string
	Logs      []TransactionLog
	GasAmount uint64
	GasPrice  *big.Int
	ShardID   uint
	ToShardID uint
}

// TokenTransaction contains a decoded transfer event, the hash of the transaction that caused the transfer
type TokenTransaction struct {
	TxHash   string
	LogIndex int
	Sender   Address
	Receiver Address
	Token    Token
	Amount   *big.Int
}

// Swap contains a decoded UniSwap swap and the hash of the transaction that caused the swap
type Swap struct {
	TxHash    string
	InToken   Token
	OutToken  Token
	InAmount  *big.Int
	OutAmount *big.Int
}

//</editor-fold>
