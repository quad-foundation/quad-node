package account

import "github.com/quad/quad-node/common"

type Account struct {
	Balance int64                      `json:"State"`
	Address [common.AddressLength]byte `json:"Address"`
}
