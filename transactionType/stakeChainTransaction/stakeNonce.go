package transactionType

import (
	"encoding/hex"
	"fmt"
	"github.com/chainpqc/chainpqc-node/account"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"strconv"
)

type StakeChainNonceData struct {
	Recipient common.Address `json:"recipient"`
	Amount    int64          `json:"amount"`
	OptData   []byte         `json:"opt_data,omitempty"`
}

type StakeChainNonce struct {
	TxData    StakeChainNonceData     `json:"tx_data"`
	TxParam   transactionType.TxParam `json:"tx_param"`
	Hash      common.Hash             `json:"hash"`
	Signature common.Signature        `json:"signature"`
	Height    int64                   `json:"height"`
	GasPrice  int64                   `json:"gas_price"`
	GasUsage  int64                   `json:"gas_usage"`
}

func (mt StakeChainNonce) GetChain() uint8 {
	return mt.TxParam.Chain
}

func (mt StakeChainNonce) GetData() StakeChainNonceData {
	return mt.TxData
}

func (mt StakeChainNonce) GetHeight() int64 {
	return mt.Height
}

func (mt StakeChainNonce) GetHash() common.Hash {
	return mt.Hash
}

func (mt StakeChainNonce) GetParam() transactionType.TxParam {
	return mt.TxParam
}

func (mt StakeChainNonce) GetSignature() common.Signature {
	return mt.Signature
}

func (mt StakeChainNonce) GasUsageEstimate() int64 {
	return 0
}

func (mt StakeChainNonce) GetGasUsage() int64 {
	return 0
}

func (tx StakeChainNonce) GetBytesWithoutSignature() []byte {

	b := tx.TxParam.GetBytes()
	b = append(b, common.GetByteInt64(tx.Height)...)
	b = append(b, common.GetByteInt64(tx.GasPrice)...)
	b = append(b, common.GetByteInt64(tx.GasUsage)...)
	b = append(b, tx.TxData.Recipient.GetBytes()...)
	b = append(b, common.GetByteInt64(tx.TxData.Amount)...)
	b = append(b, tx.TxData.OptData...)
	return b
}

func (td StakeChainNonceData) GetString() string {
	t := "Recipient: " + td.Recipient.GetHex() + "\n"
	t += "Amount PQC: " + fmt.Sprintln(account.Int64toFloat64(td.Amount)) + "\n"
	t += "Opt Data: " + hex.EncodeToString(td.OptData) + "\n"
	return t
}

func (tx StakeChainNonce) GetString() string {
	t := "Common parameters:\n" + tx.TxParam.GetString() + "\n"
	t += "Data:\n" + tx.TxData.GetString() + "\n"
	t += "Block Height: " + strconv.FormatInt(tx.Height, 10) + "\n"
	t += "Gas Price: " + strconv.FormatInt(tx.GasPrice, 10) + "\n"
	t += "Gas Usage: " + strconv.FormatInt(tx.GasUsage, 10) + "\n"
	t += "Hash: " + tx.Hash.GetHex() + "\n"
	t += "Signature: " + tx.Signature.GetHex() + "\n"
	return t
}

func (tx StakeChainNonce) GetSenderAddress() common.Address {
	return tx.TxParam.Sender
}
