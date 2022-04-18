package harmony

import (
	"math/big"
)

//<editor-fold desc="Basic types">

// Address contains wallet or contract address in one1... format as string
// and the same address as an eth address
type Address struct {
	OneAddress string
	HexAddress string
}

//</editor-fold>

//<editor-fold desc="Smart contract related types">

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

//</editor-fold

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
	Status    int
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

//</editor-fold>

//<editor-fold desc=DeFi related types>
const (
	AddLiquidity    = "add"
	RemoveLiquidity = "rem"
)

// Swap contains a decoded UniSwap swap and the hash of the transaction that caused the swap
type Swap struct {
	TxHash    string
	InToken   Token
	OutToken  Token
	InAmount  *big.Int
	OutAmount *big.Int
	Path      []LiquidityPool
	FeeToken  string
	FeeAmount *big.Int
}

// Claim contains a decoded MasterChef style claim of Tokens
type Claim struct {
	TxHash string
	Token  Token
	Amount *big.Int
}

// LiquidityAction contains the addition or removal of tokens to a LiquidityPool
type LiquidityAction struct {
	TxHash    string
	LP        LiquidityPool
	AmountA   *big.Int
	AmountB   *big.Int
	AmountLP  *big.Int
	Direction string
}

// LiquidityPool is an abstract representation of an LP contract containing the Token itself as well as the 'contained' Tokens
type LiquidityPool struct {
	TokenA  Token
	TokenB  Token
	LpToken Token
}

// HistoricLiquidityRatio contains the liquidity reserves of a pool at a given block
type HistoricLiquidityRatio struct {
	LP       LiquidityPool
	BlockNum uint64
	ReserveA *big.Int
	ReserveB *big.Int
}

//</editor-fold>
