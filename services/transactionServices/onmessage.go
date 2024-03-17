package transactionServices

import (
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/message"
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
				}
			}

			//log.Println("No of Tx in SendToPool: ", topic, " = ",
			//	transactionsPool.PoolsTx[topic[1]].NumberOfTransactions())
		}

	default:
	}
}
