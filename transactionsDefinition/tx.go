package transactionsDefinition

import (
	"encoding/hex"
	"fmt"
	"github.com/quad/quad-node/account"
	"github.com/quad/quad-node/common"
)

type TxData struct {
	Recipient common.Address `json:"recipient"`
	Amount    int64          `json:"amount"`
	OptData   []byte         `json:"opt_data,omitempty"`
}

func (td TxData) GetString() string {
	t := "Recipient: " + td.Recipient.GetHex() + "\n"
	t += "Amount PQC: " + fmt.Sprintln(account.Int64toFloat64(td.Amount)) + "\n"
	t += "Opt Data: " + hex.EncodeToString(td.OptData) + "\n"
	return t
}

func (md TxData) GetOptData() []byte {
	return md.OptData
}
func (md TxData) GetAmount() int64 {
	return md.Amount
}
func (md TxData) GetRecipient() common.Address {
	return md.Recipient
}

func (md TxData) GetBytes() ([]byte, error) {
	b := md.Recipient.GetBytes()
	b = append(b, common.GetByteInt64(md.Amount)...)
	opt := common.BytesToLenAndBytes(md.OptData)
	b = append(b, opt...)
	return b, nil
}

func (TxData) GetFromBytes(data []byte) (TxData, []byte, error) {
	md := TxData{}
	address, err := common.BytesToAddress(data[:common.AddressLength])
	if err != nil {
		return TxData{}, []byte{}, err
	}
	md.Recipient = address
	amountBytes := data[common.AddressLength : common.AddressLength+8]
	md.Amount = common.GetInt64FromByte(amountBytes)
	opt, left, err := common.BytesWithLenToBytes(data[common.AddressLength+8:])
	if err != nil {
		return TxData{}, []byte{}, err
	}
	md.OptData = opt
	return md, left, nil
}
