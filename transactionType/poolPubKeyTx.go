package transactionType

import (
	"container/heap"
)

type PoolPubKeyTx PriorityQueue

func (pp *PoolPubKeyTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *PoolPubKeyTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*PubKeyChainTransaction))
	}
}
func (pp *PoolPubKeyTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
	queue := []AnyTransaction{}
	for i := 0; i < numberTx; i++ {
		q := heap.Pop((*PriorityQueue)(pp)).(AnyTransaction)
		if q == nil {
			break
		}
		queue = append(queue, q)
	}
	return queue
}
