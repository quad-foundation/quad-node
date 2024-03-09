package account

import (
	"fmt"
	"github.com/quad/quad-node/common"
)

type Account struct {
	Balance int64                      `json:"balance"`
	Address [common.AddressLength]byte `json:"address"`
}

func (a Account) Marshal() []byte {
	b := common.GetByteInt64(a.Balance)
	b = append(b, a.Address[:]...)
	return b
}

func (a *Account) Unmarshal(data []byte) error {
	a.Balance = common.GetInt64FromByte(data[:8])
	if len(data) != 28 {
		return fmt.Errorf("wrong number of bytes in unmarshal account")
	}
	copy(a.Address[:], data[8:])
	return nil
}
