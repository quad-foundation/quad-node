package transactionServices

import (
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/message"
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

	isValid := message.CheckMessage(amsg)
	if isValid == false {
		log.Println("message is invalid")
		panic("message is invalid")
	}
	msg = amsg.(message.TransactionsMessage)
	txn, err := msg.GetTransactionsFromBytes()
	if err != nil {
		return
	}

	switch string(msg.GetHead()) {
	case "tx":

		// need to check transactions
		for topic, v := range txn {
			for _, t := range v {
				if t.Verify() {
					transactionsPool.PoolsTx[topic[1]].AddTransaction(t)
					prefix := []byte{common.TransactionDBPrefix[0], topic[1]}
					err := t.StoreToDBPoolTx(prefix)
					if err != nil {
						log.Println(err)
					}
				} else {
					log.Println("transaction verification fails")
				}
			}
		}
	case "st":
		txn := amsg.(message.TransactionsMessage).GetTransactionsBytes()
		chain := amsg.GetChain()
		for topic, v := range txn {
			txs := []transactionsDefinition.Transaction{}
			for _, hs := range v {
				t, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionDBPrefix[:], hs)
				if err != nil {
					log.Println("cannot load transaction", err)
					continue
				}
				if t.GetChain() == chain {
					txs = append(txs, t)
				} else {
					panic("wrong transaction chain")
				}
			}
			transactionMsg, err := GenerateTransactionMsg(txs, []byte("tx"), chain, topic)
			if err != nil {
				log.Println("cannot generate transaction msg", err)
			}
			Send(addr, transactionMsg.GetBytes())
		}
	default:
	}
}
