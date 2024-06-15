package stake

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	"math"
	"time"
)

type StakingAccount struct {
	StakedBalance    int64                      `json:"staked_balance"`
	StakingRewards   int64                      `json:"staking_rewards"`
	DelegatedAccount [common.AddressLength]byte `json:"delegated_account"`
	StakingDetails   map[int64][]StakingDetail  `json:"staking_details,omitempty"` // block number as key of map
}

type StakingDetail struct {
	Amount      int64 `json:"amount"`
	Reward      int64 `json:"reward"`
	LastUpdated int64 `json:"last_updated"`
}

func Stake(account *StakingAccount, amount int64, height int64) error {
	StakingRWMutex.Lock()
	defer StakingRWMutex.Unlock()
	if amount < 0 {
		return fmt.Errorf("staked amount has to be larger than 0")
	}
	account.StakedBalance += amount
	sd := StakingDetail{
		Amount:      amount,
		Reward:      0,
		LastUpdated: time.Now().Unix(),
	}
	if ContainsKeyInt64(ExtractKeysOfList(account.StakingDetails), height) == false {
		account.StakingDetails[height] = []StakingDetail{}
	}
	account.StakingDetails[height] = append(account.StakingDetails[height], sd)
	return nil
}

func Unstake(account *StakingAccount, amount int64, height int64) error {
	StakingRWMutex.Lock()
	defer StakingRWMutex.Unlock()
	if amount < 0 {
		return fmt.Errorf("unstaked amount has to be larger than 0")
	}
	if account.StakedBalance+amount < 0 {
		return fmt.Errorf("insufficient staked balance")
	}
	account.StakedBalance -= amount
	sd := StakingDetail{
		Amount:      -amount,
		LastUpdated: time.Now().Unix(),
	}
	if ContainsKeyInt64(ExtractKeysOfList(account.StakingDetails), height) == false {
		account.StakingDetails[height] = []StakingDetail{}
	}
	account.StakingDetails[height] = append(account.StakingDetails[height], sd)
	return nil
}

func Reward(account *StakingAccount, reward int64, height int64) error {
	StakingRWMutex.Lock()
	defer StakingRWMutex.Unlock()
	if reward < 0 {
		return fmt.Errorf("reward has to be larger than 0")
	}
	account.StakingRewards += reward
	sd := StakingDetail{
		Amount:      0,
		Reward:      reward,
		LastUpdated: time.Now().Unix(),
	}
	if ContainsKeyInt64(ExtractKeysOfList(account.StakingDetails), height) == false {
		account.StakingDetails[height] = []StakingDetail{}
	}
	account.StakingDetails[height] = append(account.StakingDetails[height], sd)
	return nil
}

func WithdrawReward(account *StakingAccount, amount int64, height int64) error {
	StakingRWMutex.Lock()
	defer StakingRWMutex.Unlock()
	if amount < 0 {
		return fmt.Errorf("withdraw amount has to be larger than 0")
	}
	if account.StakingRewards+amount < 0 {
		return fmt.Errorf("insufficient rewards balance to withdraw")
	}
	account.StakedBalance -= amount
	sd := StakingDetail{
		Amount:      0,
		Reward:      -amount,
		LastUpdated: time.Now().Unix(),
	}
	if ContainsKeyInt64(ExtractKeysOfList(account.StakingDetails), height) == false {
		account.StakingDetails[height] = []StakingDetail{}
	}
	account.StakingDetails[height] = append(account.StakingDetails[height], sd)
	return nil
}

// GetStakeConfirmedFloat get amount of confirmed staked QAD in human-readable format
func (a *StakingAccount) GetBalanceConfirmedFloat() float64 {
	StakingRWMutex.RLock()
	defer StakingRWMutex.RUnlock()
	return float64(a.StakedBalance) * math.Pow10(-int(common.Decimals))
}

// Marshal converts StakingAccount to a binary format.
func (sa StakingAccount) Marshal() []byte {
	StakingRWMutex.RLock()
	defer StakingRWMutex.RUnlock()
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
	StakingRWMutex.Lock()
	defer StakingRWMutex.Unlock()
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
