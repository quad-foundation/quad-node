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

	switch string(amsg.GetHead()) {
	case "tx":
		msg = amsg.(message.TransactionsMessage)
		txn, err := msg.GetTransactionsFromBytes()
		if err != nil {
			return
		}
		// need to check transactions
		for _, v := range txn {
			for _, t := range v {
				if t.Verify() {
					transactionsPool.PoolsTx.AddTransaction(t)
					err := t.StoreToDBPoolTx(common.TransactionDBPrefix[:])
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
		for topic, v := range txn {
			txs := []transactionsDefinition.Transaction{}
			for _, hs := range v {
				t, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionDBPrefix[:], hs)
				if err != nil {
					log.Println("cannot load transaction", err)
					continue
				}
				txs = append(txs, t)
			}
			transactionMsg, err := GenerateTransactionMsg(txs, []byte("tx"), topic)
			if err != nil {
				log.Println("cannot generate transaction msg", err)
			}
			Send(addr, transactionMsg.GetBytes())
		}
	default:
	}
}
