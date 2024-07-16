package account

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	"log"
	"math"
)

type Account struct {
	Balance int64                      `json:"balance"`
	Address [common.AddressLength]byte `json:"address"`
}

func GetAccountByAddressBytes(address []byte) Account {
	AccountsRWMutex.RLock()
	defer AccountsRWMutex.RUnlock()
	addrb := [common.AddressLength]byte{}
	copy(addrb[:], address[:common.AddressLength])
	return Accounts.AllAccounts[addrb]
}

func SetAccountByAddressBytes(address []byte) Account {
	dexAccount := GetAccountByAddressBytes(address)
	if !bytes.Equal(dexAccount.Address[:], address) {
		log.Println("no account found, will be created")
		addrb := [common.AddressLength]byte{}
		copy(addrb[:], address[:common.AddressLength])
		dexAccount = Account{
			Balance: 0,
			Address: addrb,
		}
		AccountsRWMutex.Lock()
		Accounts.AllAccounts[addrb] = dexAccount
		AccountsRWMutex.Unlock()
	}
	return dexAccount
}

// GetBalanceConfirmedFloat get amount of confirmed QAD in human-readable format
func (a *Account) GetBalanceConfirmedFloat() float64 {
	return float64(a.Balance) * math.Pow10(-int(common.Decimals))
}

func (a Account) Marshal() []byte {
	b := common.GetByteInt64(a.Balance)
	b = append(b, a.Address[:]...)
	return b
}

func (a *Account) Unmarshal(data []byte) error {
	if len(data) != 28 {
		return fmt.Errorf("wrong number of bytes in unmarshal account")
	}
	a.Balance = common.GetInt64FromByte(data[:8])

	copy(a.Address[:], data[8:])
	return nil
}
