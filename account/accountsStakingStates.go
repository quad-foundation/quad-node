package account

import (
	"bytes"
	"fmt"
	"github.com/quad/quad-node/account/stake"
	"github.com/quad/quad-node/common"
	memDatabase "github.com/quad/quad-node/database"
	"log"
)

type StakingAccountsType struct {
	AllStakingAccounts map[[20]byte]stake.StakingAccount `json:"all_staking_accounts"`
}

var StakingAccounts StakingAccountsType

// Marshal converts AccountsType to a binary format.
func (at StakingAccountsType) Marshal() []byte {
	var buffer bytes.Buffer

	// Number of accounts
	accountCount := len(at.AllStakingAccounts)
	buffer.Write(common.GetByteInt64(int64(accountCount)))

	// Iterate over map and marshal each account
	for address, acc := range at.AllStakingAccounts {
		buffer.Write(address[:]) // Write address
		accb := acc.Marshal()
		buffer.Write(common.BytesToLenAndBytes(accb)) // Marshal and write account
	}

	return buffer.Bytes()
}

// Unmarshal decodes AccountsType from a binary format.
func (at *StakingAccountsType) Unmarshal(data []byte) error {
	buffer := bytes.NewBuffer(data)

	// Number of accounts
	accountCount := common.GetInt64FromByte(buffer.Next(8))

	at.AllStakingAccounts = make(map[[common.AddressLength]byte]stake.StakingAccount, accountCount)

	// Read each account
	for i := int64(0); i < accountCount; i++ {
		var address [common.AddressLength]byte
		var acc stake.StakingAccount

		// Read address
		if n, err := buffer.Read(address[:]); err != nil || n != common.AddressLength {
			return fmt.Errorf("failed to read address: %w", err)
		}

		// The rest of the data is for the StakingAccount; unmarshal it
		nb := common.GetInt32FromByte(buffer.Next(4))

		if err := acc.Unmarshal(buffer.Next(int(nb))); err != nil {
			return fmt.Errorf("failed to unmarshal account: %w", err)
		}

		at.AllStakingAccounts[address] = acc
	}

	return nil
}

func StoreStakingAccounts() error {

	k := StakingAccounts.Marshal()
	err := memDatabase.MainDB.Put(common.StakingAccountsDBPrefix[:], k[:])
	if err != nil {
		log.Println("cannot store accounts", err)
		return err
	}

	return nil
}

func LoadStakingAccounts() error {
	b, err := memDatabase.MainDB.Get(common.StakingAccountsDBPrefix[:])
	if err != nil {
		log.Println("cannot load accounts", err)
		return err
	}
	err = StakingAccounts.Unmarshal(b)
	if err != nil {
		log.Println("cannot unmarshal accounts")
		return err
	}
	return nil
}
