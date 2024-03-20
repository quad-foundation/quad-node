package account

import (
	"bytes"
	"fmt"
	"github.com/quad/quad-node/common"
	memDatabase "github.com/quad/quad-node/database"
	"log"
	"sync"
)

type AccountsType struct {
	AllAccounts map[[common.AddressLength]byte]Account `json:"all_accounts"`
}

var Accounts AccountsType
var AccountsRWMutex sync.RWMutex

// error is not checked one should do the checking before
func SetBalance(address [common.AddressLength]byte, balance int64) {
	AccountsRWMutex.Lock()
	defer AccountsRWMutex.Unlock()
	acc := Accounts.AllAccounts[address]
	acc.Balance = balance
	Accounts.AllAccounts[address] = acc
}

// error is not checked one should do the checking before
func GetBalance(address [common.AddressLength]byte) int64 {
	AccountsRWMutex.RLock()
	defer AccountsRWMutex.RUnlock()
	return Accounts.AllAccounts[address].Balance
}

// Marshal converts AccountsType to a binary format.
func (at AccountsType) Marshal() []byte {
	var buffer bytes.Buffer
	AccountsRWMutex.Lock()
	defer AccountsRWMutex.Unlock()
	// Number of accounts
	accountCount := len(at.AllAccounts)
	buffer.Write(common.GetByteInt64(int64(accountCount)))

	// Iterate over map and marshal each account
	for address, acc := range at.AllAccounts {
		buffer.Write(address[:])    // Write address
		buffer.Write(acc.Marshal()) // Marshal and write account
	}

	return buffer.Bytes()
}

// Unmarshal decodes AccountsType from a binary format.
func (at *AccountsType) Unmarshal(data []byte) error {
	buffer := bytes.NewBuffer(data)
	AccountsRWMutex.Lock()
	defer AccountsRWMutex.Unlock()
	// Number of accounts
	accountCount := common.GetInt64FromByte(buffer.Next(8))

	at.AllAccounts = make(map[[common.AddressLength]byte]Account, accountCount)

	// Read each account
	for i := int64(0); i < accountCount; i++ {
		var address [common.AddressLength]byte
		var acc Account

		// Read address
		if n, err := buffer.Read(address[:]); err != nil || n != common.AddressLength {
			return fmt.Errorf("failed to read address: %w", err)
		}

		// Account binary data
		accountData := buffer.Next(common.AddressLength + 8) // Account data length (8 bytes for balance + 20 bytes for address)
		if len(accountData) != common.AddressLength+8 {
			return fmt.Errorf("incorrect account data length: got %d, want 28", len(accountData))
		}

		if err := acc.Unmarshal(accountData); err != nil {
			return fmt.Errorf("failed to unmarshal account: %w", err)
		}

		at.AllAccounts[address] = acc
	}

	return nil
}

func StoreAccounts() error {
	k := Accounts.Marshal()
	AccountsRWMutex.RLock()
	defer AccountsRWMutex.RUnlock()
	err := memDatabase.MainDB.Put(common.AccountsDBPrefix[:], k[:])
	if err != nil {
		log.Println("cannot store accounts", err)
		return err
	}
	return nil
}

func LoadAccounts() error {
	AccountsRWMutex.Lock()
	b, err := memDatabase.MainDB.Get(common.AccountsDBPrefix[:])
	if err != nil {
		log.Println("cannot load accounts", err)
		return err
	}
	AccountsRWMutex.Unlock()
	err = Accounts.Unmarshal(b)
	if err != nil {
		log.Println("cannot unmarshal accounts")
		return err
	}
	return nil
}
