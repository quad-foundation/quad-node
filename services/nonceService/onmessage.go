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
			newBlock, err := services.CreateBlockFromNonceMessage([]transactionType.AnyTransaction{transaction}, lastBlock)
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

		case "bl": //block

		default:
		}
	}
}
