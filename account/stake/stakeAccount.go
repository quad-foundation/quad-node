package stake

import (
	"bytes"
	"github.com/quad/quad-node/common"
)

type StakingAccount struct {
	StakedBalance  int64                      `json:"staked_balance"`
	StakingRewards int64                      `json:"staking_rewards"`
	Address        [common.AddressLength]byte `json:"address"`
	StakingDetails map[string]StakingDetail   `json:"staking_details,omitempty"`
}

type StakingDetail struct {
	Amount      int64 `json:"amount"`
	Reward      int64 `json:"reward"`
	LastUpdated int64 `json:"last_updated"`
}

// Marshal converts StakingAccount to a binary format.
func (sa StakingAccount) Marshal() []byte {
	var buffer bytes.Buffer

	// StakedBalance, StakingRewards
	buffer.Write(common.GetByteInt64(sa.StakedBalance))
	buffer.Write(common.GetByteInt64(sa.StakingRewards))

	// Address length and Address
	buffer.Write(sa.Address[:])

	// StakingDetails count
	buffer.Write(common.GetByteInt64(int64(len(sa.StakingDetails))))

	// StakingDetails
	for key, detail := range sa.StakingDetails {
		keyLen := int64(len(key))
		buffer.Write(common.GetByteInt64(keyLen))
		buffer.WriteString(key)

		buffer.Write(common.GetByteInt64(detail.Amount))
		buffer.Write(common.GetByteInt64(detail.Reward))
		buffer.Write(common.GetByteInt64(detail.LastUpdated))
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
	copy(sa.Address[:], buffer.Next(common.AddressLength))

	// StakingDetails
	detailsCount := common.GetInt64FromByte(buffer.Next(8))
	sa.StakingDetails = make(map[string]StakingDetail, detailsCount)

	for i := int64(0); i < detailsCount; i++ {
		keyLen := common.GetInt64FromByte(buffer.Next(8))
		key := string(buffer.Next(int(keyLen)))

		amount := common.GetInt64FromByte(buffer.Next(8))
		reward := common.GetInt64FromByte(buffer.Next(8))
		lastUpdated := common.GetInt64FromByte(buffer.Next(8))

		sa.StakingDetails[key] = StakingDetail{
			Amount:      amount,
			Reward:      reward,
			LastUpdated: lastUpdated,
		}
	}

	return nil
}
