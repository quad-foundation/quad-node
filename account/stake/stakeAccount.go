package stake

type StakingAccount struct {
	StakedBalance  int64                     `json:"staked_balance"`
	StakingRewards int64                     `json:"staking_rewards"`
	Address        []byte                    `json:"address"`
	StakingDetails map[string]*StakingDetail `json:"staking_details,omitempty"`
}

type StakingDetail struct {
	Amount      int64 `json:"amount"`
	Reward      int64 `json:"reward"`
	LastUpdated int64 `json:"last_updated"`
}
