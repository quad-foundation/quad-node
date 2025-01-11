package pubkeys

import (
	"fmt"
	"github.com/quad-foundation/quad-node/common"
)

func AddPubKeyToAddress(pk common.PubKey, mainAddress common.Address) error {
	as, err := LoadAddresses(mainAddress)
	if err != nil {
		return err
	}
	address, err := common.PubKeyToAddress(pk.GetBytes())
	if err != nil {
		return err
	}
	as = append(as, address)
	tree, err := BuildMerkleTree(mainAddress, as, GlobalMerkleTree.DB)
	if err != nil {
		return err
	}
	for _, a := range as {
		if !tree.IsAddressInTree(a) {
			return fmt.Errorf("pubkey patricia trie fails to add pubkey")
		}
	}
	err = tree.StoreTree(mainAddress)
	if err != nil {
		return err
	}
	return nil
}

func CreateAddressFromFirstPubKey(p common.PubKey) (common.Address, error) {
	address, err := common.PubKeyToAddress(p.GetBytes())
	if err != nil {
		return common.Address{}, err
	}
	as, err := LoadAddresses(address)
	if err != nil {
		return common.Address{}, err
	}
	if len(as) > 0 {
		return common.Address{}, fmt.Errorf("there are just generated markle trie for given pubkey")
	}
	tree, err := BuildMerkleTree(address, []common.Address{address}, GlobalMerkleTree.DB)
	if err != nil {
		return common.Address{}, err
	}
	if !tree.IsAddressInTree(address) {
		return common.Address{}, fmt.Errorf("addresses patricia trie fails to initialize")
	}
	err = tree.StoreTree(address)
	if err != nil {
		return common.Address{}, err
	}
	return address, nil
}
