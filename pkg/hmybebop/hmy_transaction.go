package hmybebop

import (
	"bytes"
	"encoding/hex"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/types"
	"math/big"
	"strings"
)

func EncodeTransaction(tx *types.Transaction) (data []byte, err error) {
	hash, _ := hex.DecodeString(strings.TrimPrefix(tx.TxHash, "0x"))
	ethHash, _ := hex.DecodeString(strings.TrimPrefix(tx.EthTxHash, "0x"))
	input, _ := hex.DecodeString(strings.TrimPrefix(tx.Input, "0x"))
	logs := make([]Log, len(tx.Logs))
	for i, log := range tx.Logs {
		var topics []byte
		for _, topic := range log.Topics {
			topicBytes, _ := hex.DecodeString(strings.TrimPrefix(topic, "0x"))
			topics = append(topics, topicBytes...)
		}
		logData, _ := hex.DecodeString(strings.TrimPrefix(log.Data, "0x"))
		logs[i] = Log{
			Index: uint16(log.LogIndex),
			Address: Addr{
				One: log.Address.OneAddress,
				Hex: log.Address.HexAddress,
			},
			Topics: topics,
			Data:   logData,
		}
	}

	bTx := Transaction{
		Hash:    hash,
		EthHash: ethHash,
		Sender: Addr{
			One: tx.Sender.OneAddress,
			Hex: tx.Sender.HexAddress,
		},
		Receiver: Addr{
			One: tx.Receiver.OneAddress,
			Hex: tx.Receiver.HexAddress,
		},
		BlockNum:  uint32(tx.BlockNum),
		TimeStamp: tx.Timestamp,
		Amount:    tx.Value.Bytes(),
		Input:     input,
		Logs:      logs,
		Status:    byte(tx.Status),
		GasAmount: tx.GasAmount,
		GasPrice:  tx.GasPrice.Bytes(),
		Shard:     byte(tx.ShardID),
		ToShard:   byte(tx.ToShardID),
	}
	var buff bytes.Buffer
	err = bTx.EncodeBebop(&buff)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	data = buff.Bytes()
	return
}

func DecodeTransaction(data []byte) (tx *types.Transaction, err error) {
	bTx := Transaction{}
	err = bTx.DecodeBebop(bytes.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	hash := "0x" + hex.EncodeToString(bTx.Hash)
	ethHash := "0x" + hex.EncodeToString(bTx.EthHash)
	input := "0x" + hex.EncodeToString(bTx.Input)

	sender := types.Address{
		OneAddress: bTx.Sender.One,
		HexAddress: bTx.Sender.Hex,
	}
	receiver := types.Address{
		OneAddress: bTx.Receiver.One,
		HexAddress: bTx.Receiver.Hex,
	}

	logs := make([]types.TransactionLog, len(bTx.Logs))
	for i, log := range bTx.Logs {
		addr := types.Address{
			OneAddress: log.Address.One,
			HexAddress: log.Address.Hex,
		}
		logData := "0x" + hex.EncodeToString(log.Data)
		var topics []string
		for i := 32; i < len(log.Topics); i += 32 {
			topics = append(topics, "0x"+hex.EncodeToString(log.Topics[i-32:i-1]))
		}
		logs[i] = types.TransactionLog{
			TxHash:   hash,
			LogIndex: int(log.Index),
			Address:  addr,
			Topics:   topics,
			Data:     logData,
		}
	}

	tx = &types.Transaction{
		TxHash:    hash,
		EthTxHash: ethHash,
		Sender:    sender,
		Receiver:  receiver,
		BlockNum:  uint64(bTx.BlockNum),
		Timestamp: bTx.TimeStamp,
		Value:     new(big.Int).SetBytes(bTx.Amount),
		Method:    types.Method{},
		Input:     input,
		Logs:      logs,
		Status:    int(bTx.Status),
		GasAmount: bTx.GasAmount,
		GasPrice:  new(big.Int).SetBytes(bTx.GasPrice),
		ShardID:   uint(bTx.Shard),
		ToShardID: uint(bTx.ToShard),
	}
	return
}
