package transactionServices

import (
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/message"
	"github.com/quad-foundation/quad-node/tcpip"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/quad-foundation/quad-node/transactionsPool"
	"log"
)

func OnMessage(addr [4]byte, m []byte) {
	h := common.GetHeight()
	if tcpip.IsIPBanned(addr, h, tcpip.TransactionTopic) {
		return
	}
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
		if transactionsPool.PoolsTx.NumberOfTransactions() > common.MaxTransactionInPool {
			//log.Println("no more transactions can be accepted to the pool")
			return
		}
		// need to check transactions
		for _, v := range txn {
			for _, t := range v {
				if t.Verify() {
					if transactionsPool.PoolsTx.TransactionExists(t.Hash.GetBytes()) {
						//log.Println("transaction just exists in Pool")
						continue
					}
					if transactionsDefinition.CheckFromDBPoolTx(common.TransactionDBPrefix[:], t.Hash.GetBytes()) {
						//log.Println("transaction just exists in DB")
						continue
					}
					transactionsPool.PoolsTx.AddTransaction(t)
					err := t.StoreToDBPoolTx(common.TransactionPoolHashesDBPrefix[:])
					if err != nil {
						log.Println(err)
					}
				} else {
					transactionsPool.PoolsTx.RemoveTransactionByHash(t.Hash.GetBytes())
					log.Println("transaction verification fails")
				}
			}
		}

		Spread(addr, m)

	case "st":
		txn := amsg.(message.TransactionsMessage).GetTransactionsBytes()
		for topic, v := range txn {
			txs := []transactionsDefinition.Transaction{}
			for _, hs := range v {
				t, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionPoolHashesDBPrefix[:], hs)
				if err != nil {
					//log.Println("cannot load transaction", err)
					t, err = transactionsDefinition.LoadFromDBPoolTx(common.TransactionDBPrefix[:], hs)
					if err != nil {
						//log.Println("cannot load transaction", err)
						continue
					}
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
