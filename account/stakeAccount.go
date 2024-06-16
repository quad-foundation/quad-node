package account

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	"time"
)

type StakingAccount struct {
	StakedBalance    int64                      `json:"staked_balance"`
	StakingRewards   int64                      `json:"staking_rewards"`
	DelegatedAccount [common.AddressLength]byte `json:"delegated_account"`
	Address          [common.AddressLength]byte `json:"address"`
	StakingDetails   map[int64][]StakingDetail  `json:"staking_details,omitempty"` // block number as key of map
}

type StakingDetail struct {
	Amount      int64 `json:"amount"`
	Reward      int64 `json:"reward"`
	LastUpdated int64 `json:"last_updated"`
}

func Stake(accb []byte, amount int64, height int64, delegatedAccount int) error {
	if len(accb) != common.AddressLength {
		return fmt.Errorf("wrong address length, must be %v", common.AddressLength)
	}
	acc := GetStakingAccountByAddressBytes(accb, delegatedAccount)
	if bytes.Compare(acc.Address[:], accb) != 0 {
		copy(acc.Address[:], accb)
	}
	StakingRWMutex.Lock()
	defer StakingRWMutex.Unlock()
	if amount < 0 {
		return fmt.Errorf("staked amount has to be larger than 0")
	}

	acc.StakedBalance += amount
	sd := StakingDetail{
		Amount:      amount,
		Reward:      0,
		LastUpdated: time.Now().Unix(),
	}
	if ContainsKeyInt64(ExtractKeysOfList(acc.StakingDetails), height) == false {
		acc.StakingDetails = map[int64][]StakingDetail{}
		acc.StakingDetails[height] = []StakingDetail{}
	}
	acc.StakingDetails[height] = append(acc.StakingDetails[height], sd)

	StakingAccounts[delegatedAccount].AllStakingAccounts[acc.Address] = acc
	return nil
}

func Unstake(accb []byte, amount int64, height int64, delegatedAccount int) error {
	if len(accb) != common.AddressLength {
		return fmt.Errorf("wrong address length, must be %v", common.AddressLength)
	}
	acc := GetStakingAccountByAddressBytes(accb, delegatedAccount)
	if bytes.Compare(acc.Address[:], accb) != 0 {
		return fmt.Errorf("no account present in unstaking account")
	}
	StakingRWMutex.Lock()
	defer StakingRWMutex.Unlock()
	if amount < 0 {
		return fmt.Errorf("unstaked amount has to be larger than 0")
	}

	if acc.StakedBalance+amount < 0 {
		return fmt.Errorf("insufficient staked balance")
	}
	acc.StakedBalance -= amount
	sd := StakingDetail{
		Amount:      -amount,
		LastUpdated: time.Now().Unix(),
	}
	if ContainsKeyInt64(ExtractKeysOfList(acc.StakingDetails), height) == false {
		acc.StakingDetails = map[int64][]StakingDetail{}
		acc.StakingDetails[height] = []StakingDetail{}
	}
	acc.StakingDetails[height] = append(acc.StakingDetails[height], sd)

	StakingAccounts[delegatedAccount].AllStakingAccounts[acc.Address] = acc
	return nil
}

func Reward(accb []byte, reward int64, height int64, delegatedAccount int) error {
	if len(accb) != common.AddressLength {
		return fmt.Errorf("wrong address length, must be %v", common.AddressLength)
	}
	acc := GetStakingAccountByAddressBytes(accb, delegatedAccount)
	if bytes.Compare(acc.Address[:], accb) != 0 {
		return fmt.Errorf("no account present in rewarding account")
	}
	StakingRWMutex.Lock()
	defer StakingRWMutex.Unlock()
	if reward < 0 {
		return fmt.Errorf("reward has to be larger than 0")
	}

	acc.StakingRewards += reward
	sd := StakingDetail{
		Amount:      0,
		Reward:      reward,
		LastUpdated: time.Now().Unix(),
	}
	if ContainsKeyInt64(ExtractKeysOfList(acc.StakingDetails), height) == false {
		acc.StakingDetails = map[int64][]StakingDetail{}
		acc.StakingDetails[height] = []StakingDetail{}
	}
	acc.StakingDetails[height] = append(acc.StakingDetails[height], sd)
	StakingAccounts[delegatedAccount].AllStakingAccounts[acc.Address] = acc
	return nil
}

func WithdrawReward(accb []byte, amount int64, height int64, delegatedAccount int) error {
	if len(accb) != common.AddressLength {
		return fmt.Errorf("wrong address length, must be %v", common.AddressLength)
	}
	acc := GetStakingAccountByAddressBytes(accb, delegatedAccount)
	if bytes.Compare(acc.Address[:], accb) != 0 {
		return fmt.Errorf("no account present in withdraw rewarding")
	}
	StakingRWMutex.Lock()
	defer StakingRWMutex.Unlock()
	if amount < 0 {
		return fmt.Errorf("withdraw amount has to be larger than 0")
	}

	if acc.StakingRewards+amount < 0 {
		return fmt.Errorf("insufficient rewards balance to withdraw")
	}
	acc.StakedBalance -= amount
	sd := StakingDetail{
		Amount:      0,
		Reward:      -amount,
		LastUpdated: time.Now().Unix(),
	}
	if ContainsKeyInt64(ExtractKeysOfList(acc.StakingDetails), height) == false {
		acc.StakingDetails = map[int64][]StakingDetail{}
		acc.StakingDetails[height] = []StakingDetail{}
	}
	acc.StakingDetails[height] = append(acc.StakingDetails[height], sd)

	StakingAccounts[delegatedAccount].AllStakingAccounts[acc.Address] = acc
	return nil
}

// Marshal converts StakingAccount to a binary format.
func (sa StakingAccount) Marshal() []byte {

	var buffer bytes.Buffer

	// StakedBalance, StakingRewards
	buffer.Write(common.GetByteInt64(sa.StakedBalance))
	buffer.Write(common.GetByteInt64(sa.StakingRewards))

	// Address length and Address
	buffer.Write(sa.DelegatedAccount[:])

	// StakingDetails count
	buffer.Write(common.GetByteInt64(int64(len(sa.StakingDetails))))

	// StakingDetails
	for key, details := range sa.StakingDetails {
		buffer.Write(common.GetByteInt64(key))
		buffer.Write(common.GetByteInt64(int64(len(details))))

		for _, detail := range details {
			buffer.Write(common.GetByteInt64(detail.Amount))
			buffer.Write(common.GetByteInt64(detail.Reward))
			buffer.Write(common.GetByteInt64(detail.LastUpdated))
		}
	}

	return buffer.Bytes()
}

// Unmarshal decodes StakingAccount from a binary format.
func (sa *StakingAccount) Unmarshal(data []byte) error {

	buffer := bytes.NewBuffer(data)

	// StakedBalance, StakingRewards
	sa.StakedBalance = common.GetInt64FromByte(buffer.Next(8))
	sa.StakingRewards = common.GetInt64FromByte(buffer.Next(8))

	// Address
	copy(sa.DelegatedAccount[:], buffer.Next(common.AddressLength))

	// StakingDetails
	detailsCount := common.GetInt64FromByte(buffer.Next(8))
	sa.StakingDetails = make(map[int64][]StakingDetail, detailsCount)

	for i := int64(0); i < detailsCount; i++ {
		key := common.GetInt64FromByte(buffer.Next(8))
		detailCount := common.GetInt64FromByte(buffer.Next(8))

		details := make([]StakingDetail, detailCount)
		for j := int64(0); j < detailCount; j++ {
			amount := common.GetInt64FromByte(buffer.Next(8))
			reward := common.GetInt64FromByte(buffer.Next(8))
			lastUpdated := common.GetInt64FromByte(buffer.Next(8))

			details[j] = StakingDetail{
				Amount:      amount,
				Reward:      reward,
				LastUpdated: lastUpdated,
			}
		}

		sa.StakingDetails[key] = details
	}
	return nil
}

func GetStakedInDelegatedAccount(n int) ([]Account, float64) {
	StakingRWMutex.RLock()
	defer StakingRWMutex.RUnlock()
	sum := int64(0)
	accs := []Account{}
	for _, sa := range StakingAccounts[n].AllStakingAccounts {
		acc := Account{
			Balance: sa.StakedBalance,
			Address: sa.Address,
		}
		sum += sa.StakedBalance
		accs = append(accs, acc)
	}
	return accs, float64(sum)
}
