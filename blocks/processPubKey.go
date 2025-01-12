package blocks

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
	"github.com/quad-foundation/quad-node/pubkeys"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"log"
)

func StoreAddress(mainAddress common.Address, address common.Address) error {
	index, err := pubkeys.FindAddressForMainAddress(mainAddress, address)
	if err != nil {
		return err
	}
	if index >= 0 {
		return fmt.Errorf("address just stored before")
	}
	pk, err := LoadPubKey(address.GetBytes(), mainAddress)
	if err != nil {
		return err
	}
	err = pubkeys.AddPubKeyToAddress(*pk, mainAddress)
	if err != nil {
		return err
	}
	return nil
}

func StorePubKey(pk common.PubKey) error {
	a, err := common.PubKeyToAddress(pk.GetBytes(), pk.Primary)
	if err != nil {
		return err
	}
	if !bytes.Equal(a.GetBytes(), pk.Address.GetBytes()) {
		return fmt.Errorf("address is different in pubkey and recovered from bytes")
	}
	err = memDatabase.MainDB.Put(append(common.PubKeyDBPrefix[:], a.GetBytes()...), pk.GetBytes())
	return err
}

func StorePubKeyInPatriciaTrie(pk common.PubKey) error {
	addresses, err := pubkeys.LoadAddresses(pk.MainAddress)
	if err != nil {
		return err
	}
	if len(addresses) == 0 {
		mainAddress, err2 := pubkeys.CreateAddressFromFirstPubKey(pk)
		if err2 != nil {
			return err2
		}
		if pk.MainAddress != mainAddress {
			return fmt.Errorf("error with creation of address from first pub key")
		}
		return nil
	}
	exist := false
	for _, a := range addresses {
		if bytes.Equal(a.GetBytes(), pk.Address.GetBytes()) {
			exist = true
			break
		}
	}
	if exist {
		log.Println(" address from pub key is just stored in mainaddress of patricia trie")
		return nil
	}

	address, err := common.PubKeyToAddress(pk.GetBytes(), pk.Primary)
	if err != nil {
		return err
	}
	addresses = append(addresses, address)
	tree, err := pubkeys.BuildMerkleTree(pk.MainAddress, addresses, pubkeys.GlobalMerkleTree.DB)
	if err != nil {
		return err
	}
	for _, a := range addresses {
		if !tree.IsAddressInTree(a) {
			return fmt.Errorf("pubkey patricia trie fails to add pubkey")
		}
	}
	err = tree.StoreTree(pk.MainAddress)
	if err != nil {
		return err
	}

	return nil
}

// LoadPubKey : a - address in bytes of pubkey
func LoadPubKey(a []byte, mainAddress common.Address) (pk *common.PubKey, err error) {
	pkb, err := memDatabase.MainDB.Get(append(common.PubKeyDBPrefix[:], a...))
	if err != nil {
		return &common.PubKey{}, err
	}
	err = pk.Init(pkb, mainAddress)
	if err != nil {
		return &common.PubKey{}, err
	}
	return pk, nil
}

// ProcessBlockPubKey : store pubkey on each transaction
func ProcessBlockPubKey(block Block) error {
	for _, txh := range block.TransactionsHashes {
		t, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:], txh.GetBytes())
		if err != nil {
			return err
		}
		pk := t.TxData.Pubkey
		zeroBytes := make([]byte, common.AddressLength)
		if bytes.Equal(pk.MainAddress.GetBytes(), zeroBytes) {
			return nil
		}
		err = StorePubKey(pk)
		if err != nil {
			return err
		}
		err = StorePubKeyInPatriciaTrie(pk)
		if err != nil {
			return err
		}
	}
	return nil
}
