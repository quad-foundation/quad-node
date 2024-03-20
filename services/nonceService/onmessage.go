package nonceServices

import (
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/message"
	"github.com/quad/quad-node/services"
	"github.com/quad/quad-node/services/transactionServices"
	"github.com/quad/quad-node/statistics"
	"github.com/quad/quad-node/transactionsDefinition"
	"github.com/quad/quad-node/transactionsPool"
	"log"
	"runtime/debug"
)

func OnMessage(addr string, m []byte) {

	//log.Println("New message nonce from:", addr)
	msg := message.TransactionsMessage{}

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			log.Println("recover (nonce Msg)", r)
		}

	}()

	amsg, err := msg.GetFromBytes(m)
	if err != nil {
		panic(err)
	}

	isValid := message.CheckMessage(amsg)
	if isValid == false {
		log.Println("message is invalid")
		panic("message is invalid")
	}

	switch string(amsg.GetHead()) {
	case "nn": // nonce
		//fmt.Printf("%v", nonceTransaction)
		//var topic [2]byte
		txn, err := amsg.(message.TransactionsMessage).GetTransactionsFromBytes()
		if err != nil {
			return
		}
		nonceTransaction := map[[2]byte]transactionsDefinition.Transaction{}

		for k, v := range txn {
			nonceTransaction[k] = v[0]
		}
		var transaction transactionsDefinition.Transaction
		for _, v := range nonceTransaction {
			//topic = k
			transaction = v
			break
		}
		nonceHeight := transaction.GetHeight()
		chain := transaction.GetChain()
		if common.CheckHeight(chain, nonceHeight) == false {
			log.Println("improper height value in nonceTransaction")
			return
		}
		h := common.GetHeight()

		if nonceHeight < 1 || nonceHeight != h+1 {
			//log.Print("nonce height invalid")
			return
		}

		isValid = transaction.Verify()
		if isValid == false {
			log.Println("nonce signature is invalid")
			return
		}
		lastBlock, err := blocks.LoadBlock(h)
		if err != nil {
			panic(err)
		}
		txs := transactionsPool.PoolsTx[transaction.GetChain()].PeekTransactions(int(common.MaxTransactionsPerBlock))
		txsBytes := make([][]byte, len(txs))
		transactionsHashes := []common.Hash{}
		for _, tx := range txs {
			hash := tx.GetHash().GetBytes()
			transactionsHashes = append(transactionsHashes, tx.GetHash())
			txsBytes = append(txsBytes, hash)
		}
		merkleTrie, err := transactionsPool.BuildMerkleTree(h+1, txsBytes)
		defer merkleTrie.Destroy()
		if err != nil {
			panic("cannot build merkleTrie")
		}

		newBlock, err := services.CreateBlockFromNonceMessage([]transactionsDefinition.Transaction{transaction},
			lastBlock,
			merkleTrie,
			transactionsHashes)

		if err != nil {
			panic("Error in block creation")
		}

		if newBlock.CheckProofOfSynergy() {
			services.BroadcastBlock(newBlock)
		} else {
			//log.Println("new block is not valid")
		}

		return
	case "rb": //reject block

	case "bl": //block
		common.BlockMutex.Lock()
		defer common.BlockMutex.Unlock()
		h := common.GetHeight()
		lastBlock, err := blocks.LoadBlock(h)
		if err != nil {
			panic(err)
		}
		txnbytes := amsg.GetTransactionsBytes()
		bls := map[[2]byte]blocks.Block{}
		for k, v := range txnbytes {
			if k[0] == byte('N') {

				bls[k], err = bls[k].GetFromBytes(v[0])
				newBlock := bls[k]
				if err != nil {
					log.Println(err)
					log.Println("cannot load blocks from bytes")
					return
				}
				chain := newBlock.GetChain()
				if chain != k[1] {
					log.Println("improper chain vs topic")
					return
				}
				if newBlock.GetHeader().Height != h+1 {
					//log.Println("block of too short chain")
					return
				}
				merkleTrie, err := blocks.CheckBaseBlock(newBlock, lastBlock)
				defer merkleTrie.Destroy()
				if err != nil {
					log.Println(err)
					return
				}
				hashesMissing := blocks.IsAllTransactions(newBlock)
				if len(hashesMissing) > 0 {
					transactionServices.SendGT(addr, hashesMissing, newBlock.GetChain())
				}
				err = blocks.CheckBlockAndTransferFunds(newBlock, lastBlock, merkleTrie)
				if err != nil {
					log.Println("check transfer transactions in block fails", err)
					return
				}
				err = newBlock.StoreBlock()
				if err != nil {
					log.Println(err)
					panic("cannot store block")
				}
				common.SetHeight(h + 1)
				log.Println("New Block success -------------------------------------", h+1, "-------chain", chain)
				if statistics.GmsMutex.Mutex.TryLock() {
					defer statistics.GmsMutex.Mutex.Unlock()
					stats, _ := statistics.LoadStats()
					stats.MainStats.Heights = common.GetHeight()
					stats.MainStats.Chain = chain
					stats.MainStats.Difficulty = newBlock.BaseBlock.BaseHeader.Difficulty
					stats.MainStats.Syncing = common.IsSyncing.Load()
					stats.MainStats.TimeInterval = newBlock.BaseBlock.BlockTimeStamp - lastBlock.BaseBlock.BlockTimeStamp
					empt := transactionsDefinition.EmptyTransaction()
					ntxs := 0
					for i := uint8(0); i < 5; i++ {
						if chain == i {
							hs, _ := newBlock.GetTransactionsHashes(merkleTrie, h+1)
							stats.MainStats.Transactions[i] = len(hs)
							stats.MainStats.TransactionsSize[i] = len(hs) * len(empt.GetBytes())
							ntxs += len(hs)
						}
					}
					stats.MainStats.Tps = float32(ntxs) / float32(stats.MainStats.TimeInterval)

					for i := uint8(0); i < 5; i++ {
						nt := transactionsPool.PoolsTx[i].NumberOfTransactions()
						stats.MainStats.TransactionsPending[i] = nt
						stats.MainStats.TransactionsPendingSize[i] = nt * len(empt.GetBytes())
					}

					err = stats.MainStats.SaveStats()
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	default:
	}
}
