package account

import "github.com/quad/quad-node/common"

type Accounts struct {
	AllAccounts map[[common.AddressLength]byte]*Account `json:"all_accounts"`
}
