package account

import (
	"github.com/chainpqc/chainpqc-node/account/stake"
	"github.com/chainpqc/chainpqc-node/common"
)

type StakingAccounts struct {
	AllStakingAccounts map[[common.AddressLength]byte]*stake.StakingAccount `json:"all_staking_accounts"`
}
