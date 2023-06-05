package nonceServices

import (
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/message"
	"github.com/chainpqc/chainpqc-node/services"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"log"
)

func OnMessage(addr string, m []byte) {

	//log.Println("New message nonce from:", addr)
	msg := message.AnyTransactionsMessage{}

	defer func() {
		if r := recover(); r != nil {
			//debug.PrintStack()
			log.Println("recover (nonce Msg)", r)
		}

	}()

	amsg, err := msg.GetFromBytes(m)
	if err != nil {
		return
	}
	if err != nil {
		panic(err)
	}

	isValid := message.CheckMessage(amsg)
	if isValid == false {
		log.Println("message is invalid")
		panic("message is invalid")
	}
	msg = amsg.(message.AnyTransactionsMessage)
	txn, err := msg.GetTransactions()
	if err != nil {
		return
	}

	nonceTransaction := map[[2]byte]transactionType.AnyTransaction{}
	if len(txn) > 0 {
		switch msg.GetChain() {
		case 0:
			for k, v := range txn {
				nonceTransaction[k] = v[0].(*transactionType.MainChainTransaction)
			}
		case 1:
			for k, v := range txn {
				nonceTransaction[k] = v[0].(*transactionType.PubKeyChainTransaction)
			}
		case 2:
			for k, v := range txn {
				nonceTransaction[k] = v[0].(*transactionType.StakeChainTransaction)
			}
		case 3:
			for k, v := range txn {
				nonceTransaction[k] = v[0].(*transactionType.DexChainTransaction)
			}
		case 4:
			for k, v := range txn {
				nonceTransaction[k] = v[0].(*transactionType.ContractChainTransaction)
			}
		default:
			panic("wrong chain")

		}

		switch string(msg.GetHead()) {
		case "nn": // nonce
			//fmt.Printf("%v", nonceTransaction)
			//var topic [2]byte
			var transaction transactionType.AnyTransaction
			for _, v := range nonceTransaction {
				//topic = k
				transaction = v
				break
			}
			nonceHeight := transaction.GetHeight()
			chain := transaction.GetChain()
			if common.CheckHeight(chain, nonceHeight) == false {
				panic("Unproper hieght value in nonceTransaction")
			}
			common.HeightMutex.RLock()
			h := common.GetHeight()
			common.HeightMutex.RUnlock()

			if nonceHeight < 1 || nonceHeight != h+1 {
				panic("nonce height invalid")
			}

			isValid = transactionType.VerifyTransaction(transaction)
			if isValid == false {
				panic("nonce signature is invalid")
			}
			lastBlock, err := blocks.LoadBlock(h)
			if err != nil {
				panic(err)
			}
			newBlock, err := services.CreateBlockFromNonceMessage(transaction, lastBlock)
			if err != nil {
				panic("Error in block creation")
			}

			if newBlock.CheckProofOfSynergy() {
				h = common.GetHeight()
				if newBlock.GetBaseBlock().BaseHeader.Height == h+1 {
					log.Println("New Block success -------------------------------------", h+1, "-------chain", chain)
					common.SetHeight(h + 1)
					err := blocks.StoreBlock(newBlock)
					if err != nil {
						return
					}
					services.BroadcastBlock(newBlock)
				} else {
					log.Println("too late babe")
				}
			} else {
				log.Println("new block is not valid")
			}

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
