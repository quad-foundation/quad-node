package transactionType

import (
	"container/heap"
)

type PoolContractTx PriorityQueue

func (pp *PoolContractTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *PoolContractTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*ContractChainTransaction))
	}
}
func (pp *PoolContractTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
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
