package transactionType

import (
	"encoding/hex"
	"fmt"
	"github.com/chainpqc/chainpqc-node/account"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"strconv"
)

type ContractChainTxData struct {
	Recipient common.Address `json:"recipient"`
	Amount    int64          `json:"amount"`
	OptData   []byte         `json:"opt_data,omitempty"`
}

type ContractChainTransaction struct {
	TxData    ContractChainTxData     `json:"tx_data"`
	TxParam   transactionType.TxParam `json:"tx_param"`
	Hash      common.Hash             `json:"hash"`
	Signature common.Signature        `json:"signature"`
	Height    int64                   `json:"height"`
	GasPrice  int64                   `json:"gas_price"`
	GasUsage  int64                   `json:"gas_usage"`
}

func (mt ContractChainTransaction) GetChain() uint8 {
	return mt.TxParam.Chain
}

func (mt ContractChainTransaction) GetData() transactionType.AnyDataTransaction {
	return mt.TxData
}

func (mt ContractChainTransaction) GetParam() transactionType.TxParam {
	return mt.TxParam
}

func (mt ContractChainTransaction) GasUsageEstimate() int64 {
	return 2100
}

func (mt ContractChainTransaction) GetGasUsage() int64 {
	return 2100
}

func (td ContractChainTxData) GetString() string {
	t := "Recipient: " + td.Recipient.GetHex() + "\n"
	t += "Amount PQC: " + fmt.Sprintln(account.Int64toFloat64(td.Amount)) + "\n"
	t += "Opt Data: " + hex.EncodeToString(td.OptData) + "\n"
	return t
}

func (tx ContractChainTransaction) GetString() string {
	t := "Common parameters:\n" + tx.TxParam.GetString() + "\n"
	t += "Data:\n" + tx.TxData.GetString() + "\n"
	t += "Block Height: " + strconv.FormatInt(tx.Height, 10) + "\n"
	t += "Gas Price: " + strconv.FormatInt(tx.GasPrice, 10) + "\n"
	t += "Gas Usage: " + strconv.FormatInt(tx.GasUsage, 10) + "\n"
	t += "Hash: " + tx.Hash.GetHex() + "\n"
	t += "Signature: " + tx.Signature.GetHex() + "\n"
	return t
}

func (tx ContractChainTransaction) GetSenderAddress() common.Address {
	return tx.TxParam.Sender
}

func (mt ContractChainTransaction) GetSignature() common.Signature {
	return mt.Signature
}

func (mt ContractChainTransaction) GetHeight() int64 {
	return mt.Height
}

func (mt ContractChainTransaction) GetHash() common.Hash {
	return mt.Hash
}

func (md ContractChainTxData) GetBytes() []byte {
	b := md.Recipient.GetBytes()
	b = append(b, common.GetByteInt64(md.Amount)...)
	b = append(b, md.OptData...)
	return b
}

func (mt ContractChainTransaction) GetPrice() int64 {
	return mt.GasPrice
}

func (tx ContractChainTransaction) GetBytesWithoutSignature(withHash bool) []byte {

	b := tx.TxParam.GetBytes()
	b = append(b, tx.TxData.GetBytes()...)
	b = append(b, common.GetByteInt64(tx.Height)...)
	b = append(b, common.GetByteInt64(tx.GasPrice)...)
	b = append(b, common.GetByteInt64(tx.GasUsage)...)
	if withHash {
		b = append(b, tx.GetHash().GetBytes()...)
	}
	return b
}

func (mt ContractChainTransaction) CalcHash() (common.Hash, error) {
	b := mt.GetBytesWithoutSignature(false)
	hash, err := common.GetHashFromBytes(b)
	if err != nil {
		return common.Hash{}, err
	}
	return hash, nil
}

func (mt *ContractChainTransaction) SetHash(h common.Hash) {
	mt.Hash = h
}
