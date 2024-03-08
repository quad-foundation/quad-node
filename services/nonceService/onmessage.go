package nonceServices

import (
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
	memDatabase "github.com/quad/quad-node/database"
	"github.com/quad/quad-node/message"
	"github.com/quad/quad-node/services"
	"github.com/quad/quad-node/statistics"
	"github.com/quad/quad-node/transactionsDefinition"
	"github.com/quad/quad-node/transactionsPool"
	"log"
)

func OnMessage(addr string, m []byte) {

	//log.Println("New message nonce from:", addr)
	msg := message.TransactionsMessage{}

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
	txn, err := amsg.GetTransactions()
	if err != nil {
		return
	}

	nonceTransaction := map[[2]byte]transactionsDefinition.Transaction{}

	for k, v := range txn {
		nonceTransaction[k] = v[0]
	}
	if len(txn) > 0 {
		switch string(amsg.GetHead()) {
		case "nn": // nonce
			//fmt.Printf("%v", nonceTransaction)
			//var topic [2]byte
			var transaction transactionsDefinition.Transaction
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
			h := common.GetHeight()

			if nonceHeight < 1 || nonceHeight != h+1 {
				panic("nonce height invalid")
			}

			isValid = transaction.Verify()
			if isValid == false {
				panic("nonce signature is invalid")
			}
			lastBlock, err := blocks.LoadBlock(h)
			if err != nil {
				panic(err)
			}
			txs := transactionsDefinition.PoolsTx[transaction.GetChain()].PeekTransactions(int(common.MaxTransactionsPerBlock))
			txsBytes := make([][]byte, len(txs))
			for _, tx := range txs {
				hash := tx.GetHash().GetBytes()
				txsBytes = append(txsBytes, hash)
			}
			merkleTrie, err := transactionsPool.BuildMerkleTree(h+1, txsBytes)
			defer merkleTrie.Destroy()

			newBlock, err := services.CreateBlockFromNonceMessage([]transactionsDefinition.Transaction{transaction},
				lastBlock,
				merkleTrie)

			if err != nil {
				panic("Error in block creation")
			}

			if newBlock.CheckProofOfSynergy() {
				common.BlockMutex.Lock()
				defer common.BlockMutex.Unlock()
				h = common.GetHeight()
				if newBlock.GetBaseBlock().BaseHeader.Height == h+1 {
					log.Println("New Block success -------------------------------------", h+1, "-------chain", chain)

					hashes, err := newBlock.GetTransactionsHashes(merkleTrie, h+1)
					if err != nil {
						panic(err)
					}

					log.Println("Number of transactions in block: ", len(hashes))
					txs := []transactionsDefinition.Transaction{}
					txshb := [][]byte{}
					for _, h := range hashes {
						tx := transactionsDefinition.PoolsTx[chain].PopTransactionByHash(h.GetBytes())
						txs = append(txs, tx)
						txshb = append(txshb, tx.GetHash().GetBytes())
						err = memDatabase.MainDB.Put(tx.GetHash().GetBytes(), tx.GetBytes())
						if err != nil {
							log.Println("Transaction not saved")
						}
					}
					err = merkleTrie.StoreTree(newBlock.GetBaseBlock().BaseHeader.Height, txshb)
					if err != nil {
						panic(err)
					}
					err = newBlock.StoreBlock()
					if err != nil {
						panic(err)
					}
					common.SetHeight(newBlock.GetBaseBlock().BaseHeader.Height)
					stats, _ := statistics.LoadStats()
					//stats.mutex.Lock()
					//defer stats.mutex.Unlock()
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
					err = stats.MainStats.SaveStats()
					if err != nil {
						log.Println(err)
					}
					//services.BroadcastBlock(newBlock)
					return
				} else {
					log.Println("too late babe")
				}
			} else {
				log.Println("new block is not valid")
			}

			return
		case "rb": //reject block

		case "bl": //block

		default:
		}
	}
}
