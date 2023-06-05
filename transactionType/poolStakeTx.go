package transactionType

import (
	"container/heap"
)

type PoolStakeTx PriorityQueue

func (pp *PoolStakeTx) IsEmpty() bool {
	return len(*pp) == 0
}
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
		if pp.IsEmpty() {
			break
		}
		q := heap.Pop((*PriorityQueue)(pp)).(AnyTransaction)
		queue = append(queue, q)
	}
	return queue
}

type ToSendPoolStakeTx PriorityQueue

func (pp *ToSendPoolStakeTx) IsEmpty() bool {
	return len(*pp) == 0
}
func (pp *ToSendPoolStakeTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *ToSendPoolStakeTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*StakeChainTransaction))
	}
}
func (pp *ToSendPoolStakeTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
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
