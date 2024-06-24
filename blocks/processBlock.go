package blocks

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
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
			hashes = append(hashes, hash)
		}
	}
	return hashes
}

func CheckBlockTransfers(block Block, lastBlock Block) (int64, int64, error) {
	txs := block.TransactionsHashes
	lastSupply := lastBlock.GetBlockSupply()
	accounts := map[[common.AddressLength]byte]account.Account{}
	stakingAccounts := map[[common.AddressLength]byte]account.StakingAccount{}
	totalFee := int64(0)
	for _, tx := range txs {
		hash := tx.GetBytes()
		poolTx, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:], hash)
		if err != nil {
			return 0, 0, err
		}
		err = checkTransactionInDBAndInMarkleTrie(hash)
		if err != nil {
			return 0, 0, err
		}
		fee := poolTx.GasPrice * poolTx.GasUsage
		totalFee += fee
		amount := poolTx.TxData.Amount
		total_amount := fee + amount
		address := poolTx.GetSenderAddress()
		recipientAddress := poolTx.TxData.Recipient
		n, err := account.IntDelegatedAccountFromAddress(recipientAddress)
		if err == nil && n < 512 { // delegated account
			stakingAcc := account.GetStakingAccountByAddressBytes(address.GetBytes(), n%256)
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
				transactionsDefinition.RemoveTransactionFromDBbyHash(common.TransactionPoolHashesDBPrefix[:], poolTx.Hash.GetBytes())
				return 0, 0, fmt.Errorf("staking transactions checking fails")
			}
		}
		acc := account.GetAccountByAddressBytes(address.GetBytes())
		if bytes.Compare(acc.Address[:], address.GetBytes()) != 0 {
			// remove bad transaction from pool
			transactionsPool.PoolsTx.RemoveTransactionByHash(poolTx.Hash.GetBytes())
			transactionsDefinition.RemoveTransactionFromDBbyHash(common.TransactionPoolHashesDBPrefix[:], poolTx.Hash.GetBytes())
			return 0, 0, fmt.Errorf("no account found in check block transafer")
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
			transactionsDefinition.RemoveTransactionFromDBbyHash(common.TransactionPoolHashesDBPrefix[:], poolTx.Hash.GetBytes())
			return 0, 0, fmt.Errorf("not enough funds on account")
		}

	}
	reward := account.GetReward(lastSupply)
	lastSupply += reward
	if lastSupply != block.GetBlockSupply() {
		return 0, 0, fmt.Errorf("block supply checking fails")
	}
	staked, rewarded := GetSupplyInStakedAccounts()
	coinsInDex := account.GetCoinLiquidityInDex()
	if GetSupplyInAccounts()+staked+rewarded+reward+lastBlock.BlockFee-coinsInDex != block.GetBlockSupply() {
		return 0, 0, fmt.Errorf("block supply checking fails vs account balances")
	}
	return reward, totalFee, nil
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

func checkTransactionInDBAndInMarkleTrie(hash []byte) error {
	if transactionsDefinition.CheckFromDBPoolTx(common.TransactionDBPrefix[:], hash) {
		dbTx, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionDBPrefix[:], hash)
		if err != nil {
			return err
		}
		h := dbTx.Height
		txHeight, err := transactionsPool.FindTransactionInBlocks(hash, h)
		if err != nil {
			return err
		}
		if txHeight < 0 {
			log.Println("transaction not in merkle tree. removing transaction")
			err = transactionsDefinition.RemoveTransactionFromDBbyHash(common.TransactionDBPrefix[:], hash)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("transaction was previously added in chain")
		}
	}
	return nil
}

func ProcessBlockTransfers(block Block, reward int64) error {
	txs := block.TransactionsHashes
	for _, tx := range txs {
		hash := tx.GetBytes()
		err := checkTransactionInDBAndInMarkleTrie(hash)
		if err != nil {
			return err
		}
		poolTx, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:], hash)
		if err != nil {
			return err
		}

		if poolTx.Height > block.GetHeader().Height {
			transactionsPool.PoolsTx.RemoveTransactionByHash(poolTx.Hash.GetBytes())
			return fmt.Errorf("transaction height is wrong")
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
	staked, sum, _ := account.GetStakedInDelegatedAccount(n)
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
		transactionsDefinition.RemoveTransactionFromDBbyHash(common.TransactionPoolHashesDBPrefix[:], hash)
	}
}

func EvaluateSmartContracts(bl *Block) bool {
	height := (*bl).GetHeader().Height
	if ok, logs, addresses, codes, _ := EvaluateSCForBlock(*bl); ok {
		State.SetSnapShotNum(height, State.Snapshot())
		for th, a := range addresses {

			prefix := common.OutputLogsHashesDBPrefix[:]
			err := memDatabase.MainDB.Put(append(prefix, th[:]...), []byte(logs[th]))
			if err != nil {
				log.Println("Cannot store output logs")
				return false
			}

			aa := [common.AddressLength]byte{}
			copy(aa[:], a.GetBytes())
			prefix = common.OutputAddressesHashesDBPrefix[:]
			err = memDatabase.MainDB.Put(append(prefix, th[:]...), codes[aa])
			if err != nil {
				log.Println("Cannot store address codes")
				return false
			}
		}

	} else {
		log.Println("Evaluating Smart Contract fails")
		return false
	}
	return true
}

func CheckBlockAndTransferFunds(newBlock *Block, lastBlock Block, merkleTrie *transactionsPool.MerkleTree) error {

	defer RemoveAllTransactionsRelatedToBlock(*newBlock)
	n, err := account.IntDelegatedAccountFromAddress(newBlock.GetHeader().DelegatedAccount)
	if err != nil || n < 1 || n > 255 {
		return fmt.Errorf("wrong delegated account")
	}
	opAccBlockAddr := newBlock.GetHeader().OperatorAccount
	if _, sumStaked, opAcc := account.GetStakedInDelegatedAccount(n); int64(sumStaked) < common.MinStakingForNode || bytes.Compare(opAcc.Address[:], opAccBlockAddr.GetBytes()) != 0 {
		return fmt.Errorf("not enough staked coins to be a node or not valid operetional account")
	}

	if EvaluateSmartContracts(newBlock) == false {
		return fmt.Errorf("evaluation of smart contracts in block fails")
	}

	reward, totalFee, err := CheckBlockTransfers(*newBlock, lastBlock)
	if err != nil {
		return err
	}
	newBlock.BlockFee = totalFee + lastBlock.BlockFee

	hashes := newBlock.GetBlockTransactionsHashes()
	log.Println("Number of transactions in block: ", len(hashes))

	err = ProcessBlockPubKey(*newBlock)
	if err != nil {
		return err
	}
	err = merkleTrie.StoreTree(newBlock.GetHeader().Height)
	if err != nil {
		return err
	}
	err = ProcessBlockTransfers(*newBlock, reward)
	if err != nil {
		return err
	}
	for _, h := range hashes {
		tx, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:], h.GetBytes())
		if err != nil {
			log.Println(err)
			continue
		}
		err = tx.StoreToDBPoolTx(common.TransactionDBPrefix[:])
		if err != nil {
			return err
		}
		transactionsPool.PoolsTx.RemoveTransactionByHash(h.GetBytes())
		err = tx.RemoveFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:])
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
