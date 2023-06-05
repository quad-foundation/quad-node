package transactionType

import (
	"container/heap"
)

type PoolMainTx PriorityQueue

func (pp *PoolMainTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *PoolMainTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*MainChainTransaction))
	}
}
func (pp *PoolMainTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
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
