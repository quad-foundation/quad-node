package transactionType

import (
	"container/heap"
)

type PoolStakeTx PriorityQueue

func (pp *PoolStakeTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *PoolStakeTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*StakeChainTransaction))
	}
}
func (pp *PoolStakeTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
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
