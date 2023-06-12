package account

import "github.com/chainpqc/chainpqc-node/common"

type Accounts struct {
	AllAccounts map[[common.AddressLength]byte]*Account `json:"all_accounts"`
}
