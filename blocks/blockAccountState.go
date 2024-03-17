package blocks

import (
	"fmt"
	"github.com/quad/quad-node/account"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/genesis"
)

func AddBalance(address [common.AddressLength]byte, addedAmount int64) error {
	balance := int64(0)
	if IsInKeysOfMapAccounts(account.Accounts.AllAccounts, address) {
		balance = account.Accounts.AllAccounts[address].Balance
	}
	if balance+addedAmount < 0 {
		return fmt.Errorf("Not enough funds on account")
	}
	balance += addedAmount
	account.SetBalance(address, balance)
	return nil
}

func GetSupplyInAccounts() int64 {
	sum := int64(0)
	for _, acc := range account.Accounts.AllAccounts {
		sum += acc.Balance
	}
	return sum
}

func ResetAccountsAndBlocksSync(height int64) {
	common.IsSyncing.Store(true)
	account.Accounts.AllAccounts = map[[20]byte]account.Account{}
	genesis.InitGenesis()
}
