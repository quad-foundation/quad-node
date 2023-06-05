package transactionType

import (
	"container/heap"
)

type PoolDexTx PriorityQueue

func (pp *PoolDexTx) IsEmpty() bool {
	return len(*pp) == 0
}
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
		if pp.IsEmpty() {
			break
		}
		q := heap.Pop((*PriorityQueue)(pp)).(AnyTransaction)
		queue = append(queue, q)
	}
	return queue
}

type ToSendPoolDexTx PriorityQueue

func (pp *ToSendPoolDexTx) IsEmpty() bool {
	return len(*pp) == 0
}
func (pp *ToSendPoolDexTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *ToSendPoolDexTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*DexChainTransaction))
	}
}
func (pp *ToSendPoolDexTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
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
