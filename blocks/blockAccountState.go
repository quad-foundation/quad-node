package blocks

import (
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
)

func AddBalance(address [common.AddressLength]byte, addedAmount int64) error {
	balance := int64(0)
	account.AccountsRWMutex.Lock()

	if IsInKeysOfMapAccounts(account.Accounts.AllAccounts, address) {
		balance = account.Accounts.AllAccounts[address].Balance
	} else {
		acc := account.Account{}
		acc.Balance = balance
		acc.Address = address
		account.Accounts.AllAccounts[address] = acc
	}
	if balance+addedAmount < 0 {
		account.AccountsRWMutex.Unlock()
		return fmt.Errorf("Not enough funds on account")
	}
	balance += addedAmount
	account.AccountsRWMutex.Unlock()
	account.SetBalance(address, balance)
	return nil
}

func GetSupplyInAccounts() int64 {
	sum := int64(0)
	account.AccountsRWMutex.RLock()
	defer account.AccountsRWMutex.RUnlock()
	for _, acc := range account.Accounts.AllAccounts {
		sum += acc.Balance
	}
	return sum
}
