package transactionType

import (
	"encoding/hex"
	"fmt"
	"github.com/chainpqc/chainpqc-node/account"
	"github.com/chainpqc/chainpqc-node/common"
	transactionType "github.com/chainpqc/chainpqc-node/transactionType"
	"strconv"
)

type PubKeyChainNonceData struct {
	Recipient common.PubKey `json:"recipient"`
	Amount    int64         `json:"amount"`
	OptData   []byte        `json:"opt_data,omitempty"`
}

type PubKeyChainNonce struct {
	TxData    PubKeyChainNonceData    `json:"tx_data"`
	TxParam   transactionType.TxParam `json:"tx_param"`
	Hash      common.Hash             `json:"hash"`
	Signature common.Signature        `json:"signature"`
	Height    int64                   `json:"height"`
	GasPrice  int64                   `json:"gas_price"`
	GasUsage  int64                   `json:"gas_usage"`
}

func (mt PubKeyChainNonce) GetChain() uint8 {
	return mt.TxParam.Chain
}

func (mt PubKeyChainNonce) GetData() PubKeyChainNonceData {
	return mt.TxData
}

func (mt PubKeyChainNonce) GetParam() transactionType.TxParam {
	return mt.TxParam
}

func (mt PubKeyChainNonce) GasUsageEstimate() int64 {
	return 0
}

func (mt PubKeyChainNonce) GetGasUsage() int64 {
	return 0
}

func (mt PubKeyChainNonce) GetSignature() common.Signature {
	return mt.Signature
}

func (mt PubKeyChainNonce) GetHeight() int64 {
	return mt.Height
}

func (mt PubKeyChainNonce) GetHash() common.Hash {
	return mt.Hash
}

func (tx PubKeyChainNonce) GetBytesWithoutSignature() []byte {

	b := tx.TxParam.GetBytes()
	b = append(b, common.GetByteInt64(tx.Height)...)
	b = append(b, common.GetByteInt64(tx.GasPrice)...)
	b = append(b, common.GetByteInt64(tx.GasUsage)...)
	b = append(b, tx.TxData.Recipient.GetBytes()...)
	b = append(b, tx.TxData.OptData...)
	return b
}

func (td PubKeyChainNonceData) GetString() string {
	t := "Recipient " + td.Recipient.GetHex() + "...\n"
	t += "Amount PQC: " + fmt.Sprintln(account.Int64toFloat64(td.Amount)) + "\n"
	t += "Opt Data: " + hex.EncodeToString(td.OptData) + "\n"
	return t
}

func (tx PubKeyChainNonce) GetString() string {
	t := "Common parameters:\n" + tx.TxParam.GetString() + "\n"
	t += "Data:\n" + tx.TxData.GetString() + "\n"
	t += "Block Height: " + strconv.FormatInt(tx.Height, 10) + "\n"
	t += "Gas Price: " + strconv.FormatInt(tx.GasPrice, 10) + "\n"
	t += "Gas Usage: " + strconv.FormatInt(tx.GasUsage, 10) + "\n"
	t += "Hash: " + tx.Hash.GetHex() + "\n"
	t += "Signature: " + tx.Signature.GetHex() + "\n"
	return t
}

func (tx PubKeyChainNonce) GetSenderAddress() common.Address {
	return tx.TxParam.Sender
}
