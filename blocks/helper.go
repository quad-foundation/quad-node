package blocks

import (
	"bytes"
	"fmt"
	"github.com/quad/quad-node/account"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/transactionsDefinition"
	"github.com/quad/quad-node/transactionsPool"
)

func CheckBaseBlock(newBlock Block, lastBlock Block) (*transactionsPool.MerkleTree, error) {
	chain := newBlock.GetChain()
	blockHeight := newBlock.GetHeader().Height
	if newBlock.GetBlockSupply() > common.MaxTotalSupply {
		return nil, fmt.Errorf("supply is too high")
	}
	if common.CheckHeight(chain, blockHeight) == false {
		return nil, fmt.Errorf("improper height value in block")
	}

	if bytes.Compare(lastBlock.BlockHash.GetBytes(), newBlock.GetHeader().PreviousHash.GetBytes()) != 0 {
		return nil, fmt.Errorf("last block hash not match to one stored in new block")
	}
	// needs to check block and process
	if newBlock.CheckProofOfSynergy() == false {
		return nil, fmt.Errorf("proof of synergy fails of block")
	}
	hash, err := newBlock.CalcBlockHash()
	if err != nil {
		return nil, err
	}
	if bytes.Compare(hash.GetBytes(), newBlock.BlockHash.GetBytes()) != 0 {
		return nil, fmt.Errorf("wrong hash of block")
	}
	rootMerkleTrie := newBlock.GetHeader().RootMerkleTree
	txs := newBlock.TransactionsHashes
	txsBytes := make([][]byte, len(txs))
	for _, tx := range txs {
		hash := tx.GetBytes()
		txsBytes = append(txsBytes, hash)
	}
	merkleTrie, err := transactionsPool.BuildMerkleTree(blockHeight, txsBytes)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(merkleTrie.GetRootHash(), rootMerkleTrie.GetBytes()) != 0 {
		return nil, fmt.Errorf("root merkleTrie hash check fails")
	}
	return merkleTrie, nil
}

func CheckBlockTransfers(block Block, lastBlock Block) (int64, error) {
	txs := block.TransactionsHashes
	chain := block.Chain
	lastSupply := lastBlock.GetBlockSupply()
	accounts := map[[common.AddressLength]byte]account.Account{}
	totalFee := int64(0)
	for _, tx := range txs {
		hash := tx.GetBytes()
		prefix := []byte{common.TransactionDBPrefix[0], chain}
		poolTx, err := transactionsDefinition.LoadFromDBPoolTx(prefix, hash)
		if err != nil {
			return 0, err
		}
		fee := poolTx.GasPrice * poolTx.GasUsage
		totalFee += fee
		amount := poolTx.TxData.Amount
		total_amount := fee + amount
		address := poolTx.GetSenderAddress()
		acc := account.GetAccountByAddressBytes(address.GetBytes())
		if bytes.Compare(acc.Address[:], address.GetBytes()) != 0 {
			return 0, fmt.Errorf("no account found in check block transafer")
		}
		if IsInKeysOfMapAccounts(accounts, acc.Address) {
			acc = accounts[acc.Address]
			acc.Balance -= total_amount
			accounts[acc.Address] = acc
		} else {
			acc.Balance -= total_amount
			accounts[acc.Address] = acc
		}
		if acc.Balance < 0 {
			return 0, fmt.Errorf("not enough funds on account")
		}
	}
	lastSupply -= totalFee
	if lastSupply != block.GetBlockSupply() {
		return 0, fmt.Errorf("block supply checking fails")
	}
	return totalFee, nil
}

func ExtractKeysFromMapAccounts(m map[[common.AddressLength]byte]account.Account) [][common.AddressLength]byte {
	keys := [][common.AddressLength]byte{}
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}

func IsInKeysOfMapAccounts(m map[[common.AddressLength]byte]account.Account, searchKey [common.AddressLength]byte) bool {
	keys := ExtractKeysFromMapAccounts(m)
	return common.ContainsKeyInMap(keys, searchKey)
}

func ProcessBlockTransfers(block Block) error {
	txs := block.TransactionsHashes
	chain := block.Chain
	accounts := map[[common.AddressLength]byte]account.Account{}
	recipients := map[[common.AddressLength]byte]account.Account{}
	totalFee := int64(0)
	for _, tx := range txs {
		hash := tx.GetBytes()
		prefix := []byte{common.TransactionDBPrefix[0], chain}
		poolTx, err := transactionsDefinition.LoadFromDBPoolTx(prefix, hash)
		if err != nil {
			return err
		}
		fee := poolTx.GasPrice * poolTx.GasUsage
		totalFee += fee
		amount := poolTx.TxData.Amount
		total_amount := fee + amount
		address := poolTx.GetSenderAddress()
		acc := account.GetAccountByAddressBytes(address.GetBytes())
		if bytes.Compare(acc.Address[:], address.GetBytes()) != 0 {
			return fmt.Errorf("no account found in check block transafer")
		}
		addressRecipient := poolTx.TxData.Recipient
		accRecipient := account.GetAccountByAddressBytes(addressRecipient.GetBytes())
		if bytes.Compare(accRecipient.Address[:], addressRecipient.GetBytes()) != 0 {
			return fmt.Errorf("no account found in check block transafer")
		}
		if IsInKeysOfMapAccounts(accounts, acc.Address) {
			acc = accounts[acc.Address]
			acc.Balance -= total_amount
			accounts[acc.Address] = acc
		} else {
			acc.Balance -= total_amount
			accounts[acc.Address] = acc
		}
		if acc.Balance < 0 {
			return fmt.Errorf("not enough funds on account")
		}
		if IsInKeysOfMapAccounts(recipients, accRecipient.Address) {
			accRecipient = recipients[accRecipient.Address]
			accRecipient.Balance += amount
			recipients[accRecipient.Address] = accRecipient
		} else {
			accRecipient.Balance += amount
			recipients[accRecipient.Address] = accRecipient
		}
	}
	return nil
}
