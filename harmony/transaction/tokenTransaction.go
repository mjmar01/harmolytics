package transaction

import (
	"harmolytics/harmony"
	"harmolytics/harmony/hex"
)

const (
	transferEvent = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

// DecodeTokenTransaction uses transaction logs to identify token transfers.
// This returns a list of all harmony.TokenTransaction that happened in the given harmony.Transaction.
func DecodeTokenTransaction(tx harmony.Transaction) (tTxs []harmony.TokenTransaction, err error) {
	for _, txLog := range tx.Logs {
		// Check if this is a transfer and make sure it isn't a NFT transfer
		if txLog.Topics[0] == transferEvent && txLog.Data != "0x" {
			sender, err := hex.ReadAddress(txLog.Topics[1], 0)
			if err != nil {
				return nil, err
			}
			receiver, err := hex.ReadAddress(txLog.Topics[2], 0)
			if err != nil {
				return nil, err
			}
			amount, err := hex.ReadInt(txLog.Data, 0)
			if err != nil {
				return nil, err
			}
			tTxs = append(tTxs, harmony.TokenTransaction{
				TxHash:   tx.TxHash,
				LogIndex: txLog.LogIndex,
				Sender:   sender,
				Receiver: receiver,
				Token:    harmony.Token{Address: txLog.Address},
				Amount:   amount,
			})
		}
	}
	return
}
