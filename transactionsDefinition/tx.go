package transactionsDefinition

import (
	"encoding/hex"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
)

type TxData struct {
	Recipient common.Address `json:"recipient"`
	Amount    int64          `json:"amount"`
	OptData   []byte         `json:"opt_data,omitempty"`
	Pubkey    common.PubKey  `json:"pubkey,omitempty"`
}

func (td TxData) GetString() string {
	t := "Recipient: " + td.Recipient.GetHex() + "\n"
	t += "Amount QAD: " + fmt.Sprintln(account.Int64toFloat64(td.Amount)) + "\n"
	t += "Opt Data: " + hex.EncodeToString(td.OptData) + "\n"
	if td.Pubkey.ByteValue != nil {
		t += "Pubkey: " + td.Pubkey.GetHex()[:20] + "\n"
	}
	t += "Address: " + td.Pubkey.Address.GetHex() + "\n"
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
func (md TxData) GetAddress() common.Address {
	return md.Pubkey.Address
}
func (md TxData) GetPubKey() common.PubKey {
	return md.Pubkey
}

func (md TxData) GetBytes() ([]byte, error) {
	b := md.Recipient.GetBytesWithPrimary()
	b = append(b, common.GetByteInt64(md.Amount)...)
	opt := common.BytesToLenAndBytes(md.OptData)
	b = append(b, opt...)
	adb := md.Pubkey.MainAddress.GetBytesWithPrimary()
	b = append(b, adb...)
	pk := common.BytesToLenAndBytes(md.Pubkey.GetBytes())
	b = append(b, pk...)
	return b, nil
}

func (TxData) GetFromBytes(data []byte) (TxData, []byte, error) {
	md := TxData{}
	address, err := common.BytesToAddress(data[:common.AddressLength+1])
	if err != nil {
		return TxData{}, nil, err
	}
	md.Recipient = address
	amountBytes := data[common.AddressLength+1 : common.AddressLength+9]
	md.Amount = common.GetInt64FromByte(amountBytes)
	opt, left, err := common.BytesWithLenToBytes(data[common.AddressLength+9:])
	if err != nil {
		return TxData{}, nil, err
	}
	md.OptData = opt
	adb := left[:common.AddressLength+1]
	mainAddress := common.Address{}
	err = mainAddress.Init(adb[:])
	if err != nil {
		return TxData{}, nil, err
	}
	pk, left, err := common.BytesWithLenToBytes(left[common.AddressLength+1:])
	if err != nil {
		return TxData{}, nil, err
	}
	err = md.Pubkey.Init(pk, mainAddress)
	if err != nil && len(pk) > 0 {
		return TxData{}, nil, err
	}
	return md, left, nil
}
