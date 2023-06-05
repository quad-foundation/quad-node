package transactionType

import (
	"container/heap"
)

type PoolPubKeyTx PriorityQueue

func (pp *PoolPubKeyTx) IsEmpty() bool {
	return len(*pp) == 0
}
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
		if pp.IsEmpty() {
			break
		}
		q := heap.Pop((*PriorityQueue)(pp)).(AnyTransaction)
		queue = append(queue, q)
	}
	return queue
}

type ToSendPoolPubKeyTx PriorityQueue

func (pp *ToSendPoolPubKeyTx) IsEmpty() bool {
	return len(*pp) == 0
}
func (pp *ToSendPoolPubKeyTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *ToSendPoolPubKeyTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*PubKeyChainTransaction))
	}
}
func (pp *ToSendPoolPubKeyTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
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
