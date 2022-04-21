package hmybebop

import (
	"bytes"
	"encoding/hex"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"math/big"
	"strings"
)

func EncodeTransaction(tx harmony.Transaction) (data []byte, err error) {
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
			Index:   uint16(log.LogIndex),
			Address: log.Address.Bytes,
			Topics:  topics,
			Data:    logData,
		}
	}

	bTx := Transaction{
		Hash:      hash,
		EthHash:   ethHash,
		Sender:    tx.Sender.Bytes,
		Receiver:  tx.Receiver.Bytes,
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

func DecodeTransaction(data []byte) (tx harmony.Transaction, err error) {
	bTx := Transaction{}
	err = bTx.DecodeBebop(bytes.NewReader(data))
	if err != nil {
		return harmony.Transaction{}, errors.Wrap(err, 0)
	}

	hash := "0x" + hex.EncodeToString(bTx.Hash)
	ethHash := "0x" + hex.EncodeToString(bTx.EthHash)
	input := "0x" + hex.EncodeToString(bTx.Input)

	sender := harmony.NewAddress("0x" + hex.EncodeToString(bTx.Sender))
	receiver := harmony.NewAddress("0x" + hex.EncodeToString(bTx.Receiver))

	logs := make([]harmony.TransactionLog, len(bTx.Logs))
	for i, log := range bTx.Logs {
		addr := harmony.NewAddress("0x" + hex.EncodeToString(log.Address))
		logData := "0x" + hex.EncodeToString(log.Data)
		var topics []string
		for i := 32; i < len(log.Topics); i += 32 {
			topics = append(topics, "0x"+hex.EncodeToString(log.Topics[i-32:i-1]))
		}
		logs[i] = harmony.TransactionLog{
			TxHash:   hash,
			LogIndex: int(log.Index),
			Address:  addr,
			Topics:   topics,
			Data:     logData,
		}
	}

	tx = harmony.Transaction{
		TxHash:    hash,
		EthTxHash: ethHash,
		Sender:    sender,
		Receiver:  receiver,
		BlockNum:  uint64(bTx.BlockNum),
		Timestamp: bTx.TimeStamp,
		Value:     new(big.Int).SetBytes(bTx.Amount),
		Method:    harmony.Method{},
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