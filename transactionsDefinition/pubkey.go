package transactionsDefinition

import (
	"github.com/quad/quad-node/common"
)

type PubkeyData struct {
	Pubkey common.PubKey `json:"pubkey"`
}

func (pkd PubkeyData) GetString() string {
	t := "Pubkey: " + pkd.Pubkey.GetHex()[:20] + "\n"
	t += "Address: " + pkd.Pubkey.Address.GetHex() + "\n"
	return t
}

func (pkd PubkeyData) GetAddress() common.Address {
	return pkd.Pubkey.Address
}

func (pkd PubkeyData) GetBytes() []byte {
	return pkd.Pubkey.GetBytes()
}

func (PubkeyData) GetFromBytes(data []byte) (PubkeyData, []byte, error) {
	pkd := PubkeyData{}
	err := pkd.Pubkey.Init(data[:common.PubKeyLength])
	if err != nil {
		return PubkeyData{}, nil, err
	}
	return pkd, data[common.PubKeyLength:], nil
}
