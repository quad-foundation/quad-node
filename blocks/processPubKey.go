package blocks

import (
	"github.com/quad/quad-node/common"
	memDatabase "github.com/quad/quad-node/database"
	"github.com/quad/quad-node/transactionsDefinition"
)

func StorePubKey(pk common.PubKey) error {
	a := pk.Address.GetBytes()
	err := memDatabase.MainDB.Put(append(common.PubKeyDBPrefix[:], a...), pk.GetBytes())
	return err
}

// LoadPubKey : a - address in bytes of pubkey
func LoadPubKey(a []byte) (pk *common.PubKey, err error) {
	pkb, err := memDatabase.MainDB.Get(append(common.PubKeyDBPrefix[:], a...))
	if err != nil {
		return &common.PubKey{}, err
	}
	err = pk.Init(pkb)
	if err != nil {
		return &common.PubKey{}, err
	}
	return pk, nil
}

// ProcessBlockPubKey : store pubkey on each transaction
func ProcessBlockPubKey(block Block) error {
	chain := block.GetChain()
	prefix := []byte{common.TransactionDBPrefix[0], chain}
	for _, txh := range block.TransactionsHashes {
		t, err := transactionsDefinition.LoadFromDBPoolTx(prefix, txh.GetBytes())
		if err != nil {
			return err
		}
		pk := t.TxData.Pubkey
		err = StorePubKey(pk)
		if err != nil {
			return err
		}
	}
	return nil
}
