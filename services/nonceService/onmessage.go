package nonceMsg

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/message"
	"github.com/chainpqc/chainpqc-node/transactionType"
	transactionType6 "github.com/chainpqc/chainpqc-node/transactionType/contractChainTransaction"
	transactionType5 "github.com/chainpqc/chainpqc-node/transactionType/dexChainTransaction"
	transactionType2 "github.com/chainpqc/chainpqc-node/transactionType/mainChainTransaction"
	transactionType3 "github.com/chainpqc/chainpqc-node/transactionType/pubKeyChainTransaction"
	transactionType4 "github.com/chainpqc/chainpqc-node/transactionType/stakeChainTransaction"
	"log"
)

func OnMessage(addr string, m []byte) {

	//log.Println("New message nonce from:", addr)
	msg := message.AnyNonceMessage{}

	defer func() {
		if r := recover(); r != nil {
			log.Println("recover (nonce Msg)", r)
		}

	}()

	err := msg.GetFromBytes(m)
	if err != nil {
		panic(err)
	}

	isValid := message.CheckMessage(&msg)
	if isValid == false {
		log.Println("message is invalid")
		panic("message is invalid")
	}
	txn := msg.GetTransactions()

	var nonceTransaction transactionType.AnyTransaction
	if len(txn) > 0 {
		switch msg.GetChain() {
		case 0:
			nonceTransaction = txn[0].(*transactionType2.MainChainTransaction)
		case 1:
			nonceTransaction = txn[0].(*transactionType3.PubKeyChainTransaction)
		case 2:
			nonceTransaction = txn[0].(*transactionType4.StakeChainTransaction)
		case 3:
			nonceTransaction = txn[0].(*transactionType5.DexChainTransaction)
		case 4:
			nonceTransaction = txn[0].(*transactionType6.ContractChainTransaction)
		default:
			panic("wrong chain")

		}

		switch string(msg.GetHead()) {
		case "nn": // nonce
			fmt.Printf("%v", nonceTransaction)
			nonceHeight := nonceTransaction.GetHeight()
			chain := nonceTransaction.GetChain()
			if common.CheckHeight(chain, nonceHeight) == false {
				panic("Unproper hieght value in nonceTransaction")
			}
			common.HeightMutex.RLock()
			h := common.GetHeight()
			common.HeightMutex.RUnlock()

			if nonceHeight <= 1 || nonceHeight != h+1 {
				panic("nonce height invalid")
				return
			}
			//rw := account.GetReward(nonceHeight)
			//common.BalanceMutex.Lock()

			//del := account.GetAccount(common.DelegatedAccount.GetByte(), true)
			//del.AddRewards(nonceHeight, rw)
			//common.BalanceMutex.Unlock()
			//if stake, ok := del.GetLastStake(nonceHeight); ok == true && stake < common.MinStakingForNode {
			//	log.Println("Not enough stake to be a node")
			//	return
			//}
			isValid = transactionType.VerifyTransaction(nonceTransaction)
			if isValid == false {
				panic("nonce signature is invalid")
			}
			//
			//var accLog = map[[common.GVMAddressLength]byte][2]int64{}
			//var ok bool
			//
			//lastHashesBlock, err := block.GetHashesBlockByHeight(h)
			//if err != nil {
			//	panic(err)
			//}
			//
			//lastBlock := lastHashesBlock.GetBlock()
			//
			//var blh block.AnyBlockHashes
			//
			//switch chain {
			//case 0:
			//
			//	txHashes := transaction.SelectTxHashesFromPoolToBlock(0, GetPriceThreshold())
			//	stakeHashes := stakingType.SelectTxHashesFromPoolToBlock()
			//
			//	common.BalanceMutex.Lock()
			//	accLog, ok = ProcessTransactionsInBlockForAccLog(accLog, txHashes, nonceHeight, true, 0)
			//	if ok == false {
			//		common.BalanceMutex.Unlock()
			//		panic("transaction processing fails 0")
			//	}
			//
			//	accLog, ok = ProcessStakesInBlockForAccLogs(accLog, stakeHashes, nonceHeight, true, true, 0)
			//	if !ok {
			//		common.BalanceMutex.Unlock()
			//		panic("staking transaction processing fails 0")
			//	}
			//	if ok, accLog = SaveAccountFromRewardInBlock(accLog, common.DelegatedAccount, nonceHeight, true); ok == false {
			//		common.BalanceMutex.Unlock()
			//		panic("saving processing fails, block rewards 0")
			//	}
			//	accLog, err = DistributeDelegatedAccountRewards(accLog, nonceHeight, true, common.DelegatedAccount)
			//	if err != nil {
			//		common.BalanceMutex.Unlock()
			//		panic(err)
			//	}
			//	hashesAccLog := CalcHHashesFromAccountStateLog(accLog)
			//	hhas := crypto.CalcHashFromAccLogs(hashesAccLog)
			//	common.BalanceMutex.Unlock()
			//
			//	bl, err := msg.CreateBlockFromNonceMain(lastBlock.(block.BlockSideChain), txHashes, hhas, stakeHashes)
			//	if err != nil {
			//		panic(err)
			//	}
			//	isValidPoW := block.CheckProofOfWork(bl)
			//	if isValidPoW == false {
			//		log.Println("Proof of Work on Main Chain is invalid. Nonce msg")
			//		panic("Proof of Work on Main Chain is invalid. Nonce msg")
			//	}
			//	log.Println("Creating new Main chain block:", bl.BaseBlock.Header.Height)
			//	blh = CreateBlockHashesFromBlock(bl, txHashes, stakeHashes, hashesAccLog)
			//
			//case 1:
			//
			//	txHashes := transaction.SelectTxHashesFromPoolToBlock(1, GetPriceThreshold())
			//
			//	common.BalanceMutex.Lock()
			//	accLog, ok = ProcessTransactionsInBlockForAccLog(accLog, txHashes, nonceHeight, true, 1)
			//	if ok == false {
			//		common.BalanceMutex.Unlock()
			//		panic("transaction processing fails 1")
			//	}
			//	if ok, accLog = SaveAccountFromRewardInBlock(accLog, common.DelegatedAccount, nonceHeight, true); ok == false {
			//		common.BalanceMutex.Unlock()
			//		panic("saving processing fails, block rewards 0")
			//	}
			//	accLog, err = DistributeDelegatedAccountRewards(accLog, nonceHeight, true, common.DelegatedAccount)
			//	if err != nil {
			//		common.BalanceMutex.Unlock()
			//		panic(err)
			//	}
			//	hashesAccLog := CalcHHashesFromAccountStateLog(accLog)
			//	hhas := crypto.CalcHashFromAccLogs(hashesAccLog)
			//	common.BalanceMutex.Unlock()
			//
			//	bl, err := msg.CreateBlockFromNonceSide(lastBlock.(block.BlockMainChain), txHashes, hhas)
			//	if err != nil {
			//		panic(err)
			//	}
			//	isValidPoW := block.CheckProofOfWork(bl)
			//	if isValidPoW == false {
			//		log.Println("Proof of Work on Side Chain is invalid. Nonce msg")
			//		return
			//	}
			//	log.Println("Creating new SIDE chain block:", bl.BaseBlock.Header.Height)
			//	blh = CreateBlockHashesFromBlock(bl, txHashes, []common.HHash{}, hashesAccLog)
			//}

			//anm, _ := GenerateBlockMessageHashes(blh)
			//utx, ustake, err := AddBlock(anm, true, lastHashesBlock, false, false)
			//if err != nil && len(utx) == 0 && len(ustake) == 0 {
			//	log.Println("block rejected before sending", h)
			//} else {
			//	BroadcastBlock(blh)
			//}
			//if addr != "127.0.0.1" && addr != tcpip.MyIP {
			//	SendNonceMsgBack(addr)
			//}
		case "rb": //reject block

			//common.BlockMutex.Lock()
			//defer common.BlockMutex.Unlock()

			//fromHeight := common.GetInt64FromByte(msg.Value["from_height"])
			//common.HeightMutex.RLock()
			//h := common.Height
			//common.HeightMutex.RUnlock()
			//log.Println("Enter reject block about the height", fromHeight, h)
			//if fromHeight > h {
			//	log.Println("no need to check. Current height smaller then checking", fromHeight, h)
			//	return
			//}
			//if fromHeight == 0 {
			//	log.Println("genesis block should not be in question")
			//	return
			//}
			//if h >= fromHeight && fromHeight >= h-2 { //todo to be changed
			//	backHeight := h - fromHeight + 2
			//	if backHeight >= h {
			//		backHeight = h - 1
			//	}
			//
			//	lgh, err := CheckAndPruneBlocks(backHeight, false, h)
			//	if err != nil {
			//		log.Println(err)
			//	} else {
			//		log.Println("reject blocks last good height", lgh)
			//		common.SyncingMutex.Lock()
			//		common.Height = lgh
			//		common.HeightMax = lgh
			//		common.SyncingMutex.Unlock()
			//	}
			//}
		case "bl": //block

			//common.HeightMutex.RLock()
			//h := common.Height
			//common.HeightMutex.RUnlock()
			//chain := common.GetChainFromHeight(h + 1)
			//hb := msg.GetHeight(chain)
			//if hb != h+1 {
			//	log.Println("wrong height of block", hb, h)
			//	return
			//}
			////common.BlockMutex.Lock()
			////defer common.BlockMutex.Unlock()
			//
			//blPrevious, err := block.GetHashesBlockByHeight(h)
			//if err != nil {
			//	panic(err)
			//}
			//
			//unknownTxh, unknownTxhStake, err := AddBlock(msg, true, blPrevious, false, false)
			//if err != nil {
			//	log.Println("Add block fails (block checking)", err)
			//	//if addr == tcpip.MyIP {
			//	//	lgh, err := CheckAndPruneBlocks(2, false, h)
			//	//	log.Println("prune blocks in getting block", lgh, err)
			//	//	if err == nil {
			//	//		common.HeightMutex.Lock()
			//	//		common.Height = lgh
			//	//		common.HeightMutex.Unlock()
			//	//	}
			//	//}
			//	mb, me := GenerateMessageGetHashes(h-1, h+1)
			//	MutexBlock.Lock()
			//	ExpectedMessage = me
			//	MutexBlock.Unlock()
			//	SendSync(mb, addr)
			//	common.IsSyncing.Store(true)
			//} else if len(unknownTxh) == 0 && len(unknownTxhStake) == 0 {
			//	unknownTxh, unknownTxhStake, err = AddBlock(msg, true, blPrevious, true, true)
			//	if err != nil {
			//		log.Println("Add block fails (block adding)", err)
			//		//if addr == tcpip.MyIP {
			//		//	lgh, err := CheckAndPruneBlocks(2, true, h)
			//		//	log.Println("prune blocks in getting block", lgh, err)
			//		//	if err == nil {
			//		//		common.HeightMutex.Lock()
			//		//		common.Height = lgh
			//		//		common.HeightMutex.Unlock()
			//		//	}
			//		//}
			//		//if addr != tcpip.MyIP {
			//		//	mb := GenerateMessageRejectBlocks(h)
			//		//	Send(addr, mb)
			//		//}
			//		return
			//	}
			//}
			//if len(unknownTxh) > 0 {
			//
			//	m := message.GenerateGetTransaction(unknownTxh, "get_transaction", chain)
			//	err := broadcast.SendTransactionMsg(addr, m)
			//	if err != nil {
			//		panic(err)
			//	}
			//
			//	log.Println("We need to get transactions from network and add to the system. Block is still pending")
			//	return
			//}
			//if len(unknownTxhStake) > 0 {
			//
			//	m := message2.GenerateGetStakingTransaction(unknownTxhStake, "get_staking")
			//
			//	err := broadcastStaking.SendStakingMsg(addr, m)
			//	if err != nil {
			//		panic(err)
			//	}
			//
			//	log.Println("We need to get transactions from network and add to the system. Block is still pending")
			//	return
			//}

		default:
		}
	}
}
