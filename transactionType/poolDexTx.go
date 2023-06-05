package transactionType

import (
	"container/heap"
)

type PoolDexTx PriorityQueue

func (pp *PoolDexTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *PoolDexTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*DexChainTransaction))
	}
}
func (pp *PoolDexTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
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
