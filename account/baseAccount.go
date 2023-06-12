package account

import "github.com/chainpqc/chainpqc-node/common"

type Account struct {
	Balance int64                      `json:"State"`
	Address [common.AddressLength]byte `json:"Address"`
}
