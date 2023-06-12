package transactionServices

import (
	"github.com/chainpqc/chainpqc-node/message"
	"github.com/chainpqc/chainpqc-node/statistics"
	"github.com/chainpqc/chainpqc-node/transactionType"
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
	txn, err := msg.GetTransactions()
	if err != nil {
		return
	}

	txs := map[[2]byte][]*transactionType.Transaction{}
	var at *transactionType.Transaction
	for k, v := range txn {
		for _, t := range v {
			at = &t
			txs[k] = append(txs[k], at)

			if len(txs) == 0 {
				return
			}
			switch string(msg.GetHead()) {
			case "tx": // nonce

				// need to check transactions
				for topic, v := range txs {
					for _, t := range v {
						transactionType.PoolsTx[topic[1]].AddTransaction(*t)
						err := t.StoreToDBPoolTx(topic[:])
						if err != nil {
							log.Println(err)
						}
					}

					log.Println("No of Tx in SendToPool: ", topic, " = ",
						transactionType.PoolsTx[topic[1]].NumberOfTransactions())
				}
				stats, _ := statistics.LoadStats()
				empt := transactionType.EmptyTransaction()
				for i := uint8(0); i < 5; i++ {
					nt := transactionType.PoolsTx[i].NumberOfTransactions()
					stats.TransactionsPending[i] = nt
					stats.TransactionsPendingSize[i] = nt * len(empt.GetBytes())
				}
				stats.SaveStats()

			default:
			}
		}
	}
}
