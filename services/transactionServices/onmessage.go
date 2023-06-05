package transactionServices

import (
	"github.com/chainpqc/chainpqc-node/message"
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

	txs := map[[2]byte][]transactionType.AnyTransaction{}
	if len(txn) > 0 {
		switch msg.GetChain() {
		case 0:
			for k, v := range txn {
				for _, t := range v {
					txs[k] = append(txs[k], t.(*transactionType.MainChainTransaction))
				}
			}
		case 1:
			for k, v := range txn {
				for _, t := range v {
					txs[k] = append(txs[k], t.(*transactionType.PubKeyChainTransaction))
				}
			}
		case 2:
			for k, v := range txn {
				for _, t := range v {
					txs[k] = append(txs[k], t.(*transactionType.StakeChainTransaction))
				}
			}
		case 3:
			for k, v := range txn {
				for _, t := range v {
					txs[k] = append(txs[k], t.(*transactionType.DexChainTransaction))
				}
			}
		case 4:
			for k, v := range txn {
				for _, t := range v {
					txs[k] = append(txs[k], t.(*transactionType.ContractChainTransaction))
				}
			}
		default:
			panic("wrong chain")

		}
		if len(txs) == 0 {
			return
		}
		switch string(msg.GetHead()) {
		case "tx": // nonce

			// need to check transactions
			for topic, v := range txs {
				transactionType.AddTransactionsToPool(v, topic[1])
			}
		default:
		}
	}
}
