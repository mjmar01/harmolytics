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
	logs := make([]log, len(tx.Logs))
	for i, l := range tx.Logs {
		var topics []byte
		for _, topic := range l.Topics {
			topicBytes, _ := hex.DecodeString(strings.TrimPrefix(topic, "0x"))
			topics = append(topics, topicBytes...)
		}
		logData, _ := hex.DecodeString(strings.TrimPrefix(l.Data, "0x"))
		logs[i] = log{
			index: uint16(l.LogIndex),
			address: addr{
				one: l.Address.OneAddress,
				hex: l.Address.HexAddress,
			},
			topics: topics,
			data:   logData,
		}
	}

	bTx := transaction{
		hash:    hash,
		ethHash: ethHash,
		sender: addr{
			one: tx.Sender.OneAddress,
			hex: tx.Sender.HexAddress,
		},
		receiver: addr{
			one: tx.Receiver.OneAddress,
			hex: tx.Receiver.HexAddress,
		},
		blockNum:  uint32(tx.BlockNum),
		timeStamp: tx.Timestamp,
		amount:    tx.Value.Bytes(),
		input:     input,
		method: method{
			signature: tx.Method.Signature,
			name:      tx.Method.Name,
			params:    tx.Method.Parameters,
		},
		logs:      logs,
		status:    byte(tx.Status),
		gasAmount: tx.GasAmount,
		gasPrice:  tx.GasPrice.Bytes(),
		shard:     byte(tx.ShardID),
		toShard:   byte(tx.ToShardID),
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
	bTx := transaction{}
	err = bTx.DecodeBebop(bytes.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	hash := "0x" + hex.EncodeToString(bTx.hash)
	ethHash := "0x" + hex.EncodeToString(bTx.ethHash)
	input := "0x" + hex.EncodeToString(bTx.input)

	sender := types.Address{
		OneAddress: bTx.sender.one,
		HexAddress: bTx.sender.hex,
	}
	receiver := types.Address{
		OneAddress: bTx.receiver.one,
		HexAddress: bTx.receiver.hex,
	}

	logs := make([]types.TransactionLog, len(bTx.logs))
	for i, l := range bTx.logs {
		a := types.Address{
			OneAddress: l.address.one,
			HexAddress: l.address.hex,
		}
		logData := "0x" + hex.EncodeToString(l.data)
		topics := make([]string, len(l.topics)/32)
		for i := 32; i <= len(l.topics); i += 32 {
			topics[(i/32)-1] = "0x" + hex.EncodeToString(l.topics[i-32:i])
		}
		logs[i] = types.TransactionLog{
			TxHash:   hash,
			LogIndex: int(l.index),
			Address:  a,
			Topics:   topics,
			Data:     logData,
		}
	}

	tx = &types.Transaction{
		TxHash:    hash,
		EthTxHash: ethHash,
		Sender:    sender,
		Receiver:  receiver,
		BlockNum:  uint64(bTx.blockNum),
		Timestamp: bTx.timeStamp,
		Value:     new(big.Int).SetBytes(bTx.amount),
		Method: types.Method{
			Signature:  bTx.method.signature,
			Name:       bTx.method.name,
			Parameters: bTx.method.params,
		},
		Input:     input,
		Logs:      logs,
		Status:    int(bTx.status),
		GasAmount: bTx.gasAmount,
		GasPrice:  new(big.Int).SetBytes(bTx.gasPrice),
		ShardID:   uint(bTx.shard),
		ToShardID: uint(bTx.toShard),
	}
	return
}

func EncodeMethod(m *types.Method) (data []byte, err error) {
	bM := method{
		signature: m.Signature,
		name:      m.Name,
		params:    m.Parameters,
	}
	var buff bytes.Buffer
	err = bM.EncodeBebop(&buff)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	data = buff.Bytes()
	return
}

func DecodeMethod(data []byte) (m *types.Method, err error) {
	bM := method{}
	err = bM.DecodeBebop(bytes.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	m = &types.Method{
		Signature:  bM.signature,
		Name:       bM.name,
		Parameters: bM.params,
	}
	return
}
