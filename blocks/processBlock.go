package blocks

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/quad-foundation/quad-node/transactionsPool"
	"log"
)

func CheckBaseBlock(newBlock Block, lastBlock Block) (*transactionsPool.MerkleTree, error) {
	blockHeight := newBlock.GetHeader().Height
	if newBlock.GetBlockSupply() > common.MaxTotalSupply {
		return nil, fmt.Errorf("supply is too high")
	}

	if bytes.Compare(lastBlock.BlockHash.GetBytes(), newBlock.GetHeader().PreviousHash.GetBytes()) != 0 {
		log.Println("lastBlock.BlockHash", lastBlock.BlockHash.GetHex(), newBlock.GetHeader().PreviousHash.GetHex())
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

func IsAllTransactions(block Block) [][]byte {
	txs := block.TransactionsHashes
	hashes := [][]byte{}
	for _, tx := range txs {
		hash := tx.GetBytes()
		isKey := transactionsDefinition.CheckFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:], hash)
		if isKey == false {
			isKeyPerm := transactionsDefinition.CheckFromDBPoolTx(common.TransactionDBPrefix[:], hash)
			if isKeyPerm {
				err := transactionsDefinition.RemoveTransactionFromDBbyHash(common.TransactionDBPrefix[:], hash)
				if err != nil {
					log.Println(err)
				}
			}
			hashes = append(hashes, hash)
		}
	}
	return hashes
}

func CheckBlockTransfers(block Block, lastBlock Block) (int64, error) {
	txs := block.TransactionsHashes
	lastSupply := lastBlock.GetBlockSupply()
	accounts := map[[common.AddressLength]byte]account.Account{}
	stakingAccounts := map[[common.AddressLength]byte]account.StakingAccount{}
	totalFee := int64(0)
	for _, tx := range txs {
		hash := tx.GetBytes()
		poolTx, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:], hash)
		if err != nil {
			return 0, err
		}
		if transactionsDefinition.CheckFromDBPoolTx(common.TransactionDBPrefix[:], poolTx.Hash.GetBytes()) {
			log.Println("transaction just exists in DB")
			continue
		}
		fee := poolTx.GasPrice * poolTx.GasUsage
		totalFee += fee
		amount := poolTx.TxData.Amount
		total_amount := fee + amount
		address := poolTx.GetSenderAddress()
		recipientAddress := poolTx.TxData.Recipient
		n, err := account.IntDelegatedAccountFromAddress(recipientAddress)
		if err == nil { // delegated account
			stakingAcc := account.GetStakingAccountByAddressBytes(address.GetBytes(), n)
			if bytes.Compare(stakingAcc.Address[:], address.GetBytes()) != 0 {
				log.Println("no account found in check block transfer")
				copy(stakingAcc.Address[:], address.GetBytes())
				copy(stakingAcc.DelegatedAccount[:], recipientAddress.GetBytes())
			}
			if IsInKeysOfMapStakingAccounts(stakingAccounts, stakingAcc.Address) {
				stakingAcc = stakingAccounts[stakingAcc.Address]
			}
			stakingAcc.StakedBalance += amount
			stakingAcc.StakingRewards += fee // just using for fee in the local copy
			stakingAccounts[stakingAcc.Address] = stakingAcc
			ret := CheckStakingTransaction(poolTx, stakingAccounts[stakingAcc.Address].StakedBalance, stakingAccounts[stakingAcc.Address].StakingRewards)
			if ret == false {
				// remove bad transaction from pool
				transactionsPool.PoolsTx.RemoveTransactionByHash(poolTx.Hash.GetBytes())
				return 0, fmt.Errorf("staking transactions checking fails")
			}
		}
		acc := account.GetAccountByAddressBytes(address.GetBytes())
		if bytes.Compare(acc.Address[:], address.GetBytes()) != 0 {
			// remove bad transaction from pool
			transactionsPool.PoolsTx.RemoveTransactionByHash(poolTx.Hash.GetBytes())
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
			// remove bad transaction from pool
			transactionsPool.PoolsTx.RemoveTransactionByHash(poolTx.Hash.GetBytes())
			return 0, fmt.Errorf("not enough funds on account")
		}
	}
	reward := account.GetReward(lastSupply)
	lastSupply += reward - totalFee
	if lastSupply != block.GetBlockSupply() {
		return 0, fmt.Errorf("block supply checking fails")
	}
	staked, rewarded := GetSupplyInStakedAccounts()
	if GetSupplyInAccounts()+staked+rewarded+reward-totalFee != block.GetBlockSupply() {
		return 0, fmt.Errorf("block supply checking fails vs account balances")
	}
	return reward, nil
}

func ExtractKeysFromMapAccounts(m map[[common.AddressLength]byte]account.Account) [][common.AddressLength]byte {
	keys := [][common.AddressLength]byte{}
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}

func ExtractKeysFromMapStakingAccounts(m map[[common.AddressLength]byte]account.StakingAccount) [][common.AddressLength]byte {
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

func IsInKeysOfMapStakingAccounts(m map[[common.AddressLength]byte]account.StakingAccount, searchKey [common.AddressLength]byte) bool {
	keys := ExtractKeysFromMapStakingAccounts(m)
	return common.ContainsKeyInMap(keys, searchKey)
}

func ProcessBlockTransfers(block Block, reward int64) error {
	txs := block.TransactionsHashes
	for _, tx := range txs {
		hash := tx.GetBytes()
		poolTx, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:], hash)
		if err != nil {
			return err
		}
		err = ProcessTransaction(poolTx, block.GetHeader().Height)
		if err != nil {
			// remove bad transaction from pool
			transactionsPool.PoolsTx.RemoveTransactionByHash(poolTx.Hash.GetBytes())
			return err
		}
	}
	addr := block.BaseBlock.BaseHeader.OperatorAccount.ByteValue
	n, err := account.IntDelegatedAccountFromAddress(block.BaseBlock.BaseHeader.DelegatedAccount)
	if err != nil || n < 1 || n > 255 {
		return fmt.Errorf("wrong delegated account in block")
	}
	staked, sum := account.GetStakedInDelegatedAccount(n)
	if sum <= 0 {
		return fmt.Errorf("no staked amount in delegated account which was rewarded")
	}
	rest := reward
	for _, acc := range staked {
		if acc.Balance > 0 {
			userReward := int64(float64(reward) * float64(acc.Balance) / sum)
			rest -= userReward // in the case when rounding lose some fraction of coins
			err := account.Reward(acc.Address[:], userReward, block.GetHeader().Height, n)
			if err != nil {
				return err
			}
		}
	}
	if rest > 0 {
		err := account.Reward(addr[:], rest, block.GetHeader().Height, n)
		if err != nil {
			return err
		}
	} else if rest < 0 {
		return fmt.Errorf("this shouldn't happen anytime")
	}

	return nil
}

func RemoveAllTransactionsRelatedToBlock(newBlock Block) {
	txs := newBlock.TransactionsHashes
	for _, tx := range txs {
		hash := tx.GetBytes()
		transactionsPool.PoolsTx.RemoveTransactionByHash(hash)
	}
}

func CheckBlockAndTransferFunds(newBlock Block, lastBlock Block, merkleTrie *transactionsPool.MerkleTree) error {

	defer RemoveAllTransactionsRelatedToBlock(newBlock)
	n, err := account.IntDelegatedAccountFromAddress(newBlock.GetHeader().DelegatedAccount)
	if err != nil || n < 1 || n > 255 {
		return fmt.Errorf("wrong delegated account")
	}
	if _, sumStaked := account.GetStakedInDelegatedAccount(n); int64(sumStaked) < common.MinStakingForNode {
		return fmt.Errorf("not enough staked coins to be a node")
	}

	reward, err := CheckBlockTransfers(newBlock, lastBlock)
	if err != nil {
		return err
	}

	hashes := newBlock.GetBlockTransactionsHashes()
	log.Println("Number of transactions in block: ", len(hashes))
	txshb := [][]byte{}
	for _, h := range hashes {
		tx := transactionsPool.PoolsTx.PopTransactionByHash(h.GetBytes())
		txshb = append(txshb, tx.GetHash().GetBytes())
		err = tx.StoreToDBPoolTx(common.TransactionDBPrefix[:])
		if err != nil {
			return err
		}
	}

	err = ProcessBlockPubKey(newBlock)
	if err != nil {
		return err
	}
	err = merkleTrie.StoreTree(newBlock.GetHeader().Height, txshb)
	if err != nil {
		return err
	}
	err = ProcessBlockTransfers(newBlock, reward)
	if err != nil {
		return err
	}
	for _, h := range hashes {
		tx, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionDBPrefix[:], h.GetBytes())
		if err != nil {
			log.Println(err)
			continue
		}
		err = tx.RemoveFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:])
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
