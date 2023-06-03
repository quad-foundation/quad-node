package transactionType

import (
	"encoding/hex"
	"fmt"
	"github.com/chainpqc/chainpqc-node/account"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"strconv"
)

type StakeChainTxData struct {
	Recipient common.Address `json:"recipient"`
	Amount    int64          `json:"amount"`
	OptData   []byte         `json:"opt_data,omitempty"`
}

type StakeChainTransaction struct {
	TxData    StakeChainTxData        `json:"tx_data"`
	TxParam   transactionType.TxParam `json:"tx_param"`
	Hash      common.Hash             `json:"hash"`
	Signature common.Signature        `json:"signature"`
	Height    int64                   `json:"height"`
	GasPrice  int64                   `json:"gas_price"`
	GasUsage  int64                   `json:"gas_usage"`
}

func (mt StakeChainTransaction) GetChain() uint8 {
	return mt.TxParam.Chain
}

func (mt StakeChainTransaction) GetData() StakeChainTxData {
	return mt.TxData
}

func (mt StakeChainTransaction) GetHeight() int64 {
	return mt.Height
}

func (mt StakeChainTransaction) GetHash() common.Hash {
	return mt.Hash
}

func (mt StakeChainTransaction) GetParam() transactionType.TxParam {
	return mt.TxParam
}

func (mt StakeChainTransaction) GetSignature() common.Signature {
	return mt.Signature
}

func (mt StakeChainTransaction) GasUsageEstimate() int64 {
	return 2100
}

func (mt StakeChainTransaction) GetGasUsage() int64 {
	return 2100
}

func (tx StakeChainTransaction) GetBytesWithoutSignature() []byte {

	b := tx.TxParam.GetBytes()
	b = append(b, common.GetByteInt64(tx.Height)...)
	b = append(b, common.GetByteInt64(tx.GasPrice)...)
	b = append(b, common.GetByteInt64(tx.GasUsage)...)
	b = append(b, tx.TxData.Recipient.GetBytes()...)
	b = append(b, common.GetByteInt64(tx.TxData.Amount)...)
	b = append(b, tx.TxData.OptData...)
	return b
}

func (td StakeChainTxData) GetString() string {
	t := "Delegated Account: " + td.Recipient.GetHex() + "\n"
	t += "Amount PQC: " + fmt.Sprintln(account.Int64toFloat64(td.Amount)) + "\n"
	t += "Opt Data: " + hex.EncodeToString(td.OptData) + "\n"
	return t
}

func (tx StakeChainTransaction) GetString() string {
	t := "Common parameters:\n" + tx.TxParam.GetString() + "\n"
	t += "Data:\n" + tx.TxData.GetString() + "\n"
	t += "Block Height: " + strconv.FormatInt(tx.Height, 10) + "\n"
	t += "Gas Price: " + strconv.FormatInt(tx.GasPrice, 10) + "\n"
	t += "Gas Usage: " + strconv.FormatInt(tx.GasUsage, 10) + "\n"
	t += "Hash: " + tx.Hash.GetHex() + "\n"
	t += "Signature: " + tx.Signature.GetHex() + "\n"
	return t
}

func (tx StakeChainTransaction) GetSenderAddress() common.Address {
	return tx.TxParam.Sender
}
