package account

import (
	"github.com/quad/quad-node/account/stake"
	"github.com/quad/quad-node/common"
)

type StakingAccounts struct {
	AllStakingAccounts map[[common.AddressLength]byte]*stake.StakingAccount `json:"all_staking_accounts"`
}
