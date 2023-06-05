package transactionType

import (
	"container/heap"
)

type PoolMainTx PriorityQueue

func (pp *PoolMainTx) IsEmpty() bool {
	return len(*pp) == 0
}
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
		if pp.IsEmpty() {
			break
		}
		q := heap.Pop((*PriorityQueue)(pp)).(AnyTransaction)
		queue = append(queue, q)
	}
	return queue
}

type ToSendPoolMainTx PriorityQueue

func (pp *ToSendPoolMainTx) IsEmpty() bool {
	return len(*pp) == 0
}
func (pp *ToSendPoolMainTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *ToSendPoolMainTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*MainChainTransaction))
	}
}
func (pp *ToSendPoolMainTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
	queue := []AnyTransaction{}
	for i := 0; i < numberTx; i++ {
		if pp.IsEmpty() {
			break
		}
		q := heap.Pop((*PriorityQueue)(pp)).(AnyTransaction)
		queue = append(queue, q)
	}
	return queue
}
