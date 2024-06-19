package account

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
	"log"
)

type StakingAccountsType struct {
	AllStakingAccounts map[[20]byte]StakingAccount `json:"all_staking_accounts"`
}

var StakingAccounts [256]StakingAccountsType

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

	at.AllStakingAccounts = make(map[[common.AddressLength]byte]StakingAccount, accountCount)

	// Read each account
	for i := int64(0); i < accountCount; i++ {
		var address [common.AddressLength]byte
		var acc StakingAccount

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

func StoreStakingAccounts(height int64) error {

	for i := 1; i < 256; i++ {
		StakingRWMutex.RLock()
		k := StakingAccounts[i].Marshal()
		StakingRWMutex.RUnlock()
		hb := common.GetByteInt64(height)
		prefix := append(common.StakingAccountsDBPrefix[:], hb...)
		prefix = append(prefix, byte(i))
		err := memDatabase.MainDB.Put(prefix, k[:])
		if err != nil {
			log.Println("cannot store accounts", err)
		}
	}
	return nil
}

func LoadStakingAccounts(height int64) error {
	var err error
	if height < 0 {
		height, err = LastHeightStoredInStakingAccounts()
		if err != nil {
			log.Println(err)
		}
	}

	for i := 1; i < 256; i++ {
		hb := common.GetByteInt64(height)
		prefix := append(common.StakingAccountsDBPrefix[:], hb...)
		prefix = append(prefix, byte(i))
		b, err := memDatabase.MainDB.Get(prefix)
		if err != nil {
			log.Println("cannot load accounts", err)
			continue
		}
		StakingRWMutex.Lock()
		err = (&StakingAccounts[i]).Unmarshal(b)
		StakingRWMutex.Unlock()
		if err != nil {
			log.Println("cannot unmarshal accounts", err)
			return err
		}
	}
	return nil
}

func GetStakingAccountByAddressBytes(address []byte, delegatedAccount int) StakingAccount {
	StakingRWMutex.RLock()
	defer StakingRWMutex.RUnlock()
	addrb := [common.AddressLength]byte{}
	copy(addrb[:], address[:common.AddressLength])
	return StakingAccounts[delegatedAccount].AllStakingAccounts[addrb]
}

func RemoveStakingAccountsFromDB(height int64) error {
	hb := common.GetByteInt64(height)
	prefix := append(common.StakingAccountsDBPrefix[:], hb...)
	for i := 1; i < 256; i++ {
		prefix2 := [2]byte{byte(i / 16), byte(i % 16)}
		prefix = append(prefix, prefix2[:]...)
		err := memDatabase.MainDB.Delete(prefix)
		if err != nil {
			log.Println("cannot remove account", err)
			return err
		}
	}
	return nil
}

func LastHeightStoredInStakingAccounts() (int64, error) {
	i := int64(0)
	for {
		ib := common.GetByteInt64(i)
		prefix := append(common.StakingAccountsDBPrefix[:], ib...)
		prefix = append(prefix, []byte{0, 1}...)
		isKey, err := memDatabase.MainDB.IsKey(prefix)
		if err != nil {
			return i - 1, err
		}
		if isKey == false {
			break
		}
		i++
	}
	return i - 1, nil
}
