package transactionType

import (
	"encoding/hex"
	"fmt"
	"github.com/chainpqc/chainpqc-node/account"
	"github.com/chainpqc/chainpqc-node/common"
	transactionType "github.com/chainpqc/chainpqc-node/transactionType"
	"strconv"
)

type PubKeyChainTxData struct {
	Recipient common.PubKey `json:"recipient"`
	Amount    int64         `json:"amount"`
	OptData   []byte        `json:"opt_data,omitempty"`
}

type PubKeyChainTransaction struct {
	TxData    PubKeyChainTxData       `json:"tx_data"`
	TxParam   transactionType.TxParam `json:"tx_param"`
	Hash      common.Hash             `json:"hash"`
	Signature common.Signature        `json:"signature"`
	Height    int64                   `json:"height"`
	GasPrice  int64                   `json:"gas_price"`
	GasUsage  int64                   `json:"gas_usage"`
}

func (mt PubKeyChainTransaction) GetChain() uint8 {
	return mt.TxParam.Chain
}

func (mt PubKeyChainTransaction) GetData() PubKeyChainTxData {
	return mt.TxData
}

func (mt PubKeyChainTransaction) GetParam() transactionType.TxParam {
	return mt.TxParam
}

func (mt PubKeyChainTransaction) GasUsageEstimate() int64 {
	return 2100
}

func (mt PubKeyChainTransaction) GetGasUsage() int64 {
	return 2100
}

func (mt PubKeyChainTransaction) GetSignature() common.Signature {
	return mt.Signature
}

func (mt PubKeyChainTransaction) GetHeight() int64 {
	return mt.Height
}

func (mt PubKeyChainTransaction) GetHash() common.Hash {
	return mt.Hash
}

func (tx PubKeyChainTransaction) GetBytesWithoutSignature() []byte {

	b := tx.TxParam.GetBytes()
	b = append(b, common.GetByteInt64(tx.Height)...)
	b = append(b, common.GetByteInt64(tx.GasPrice)...)
	b = append(b, common.GetByteInt64(tx.GasUsage)...)
	b = append(b, tx.TxData.Recipient.GetBytes()...)
	b = append(b, tx.TxData.OptData...)
	return b
}

func (td PubKeyChainTxData) GetString() string {
	t := "Recipient: " + td.Recipient.GetHex()[:100] + "...\n"
	t += "Amount PQC: " + fmt.Sprintln(account.Int64toFloat64(td.Amount)) + "\n"
	t += "Public Key: " + hex.EncodeToString(td.OptData) + "\n"
	return t
}

func (tx PubKeyChainTransaction) GetString() string {
	t := "Common parameters:\n" + tx.TxParam.GetString() + "\n"
	t += "Data:\n" + tx.TxData.GetString() + "\n"
	t += "Block Height: " + strconv.FormatInt(tx.Height, 10) + "\n"
	t += "Gas Price: " + strconv.FormatInt(tx.GasPrice, 10) + "\n"
	t += "Gas Usage: " + strconv.FormatInt(tx.GasUsage, 10) + "\n"
	t += "Hash: " + tx.Hash.GetHex() + "\n"
	t += "Signature: " + tx.Signature.GetHex() + "\n"
	return t
}

func (tx PubKeyChainTransaction) GetSenderAddress() common.Address {
	return tx.TxParam.Sender
}
