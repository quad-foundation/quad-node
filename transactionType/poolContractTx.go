package transactionType

import (
	"container/heap"
)

type PoolContractTx PriorityQueue

func (pp *PoolContractTx) IsEmpty() bool {
	return len(*pp) == 0
}
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
		if pp.IsEmpty() {
			break
		}
		q := heap.Pop((*PriorityQueue)(pp)).(AnyTransaction)
		queue = append(queue, q)
	}
	return queue
}

type ToSendPoolContractTx PriorityQueue

func (pp *ToSendPoolContractTx) IsEmpty() bool {
	return len(*pp) == 0
}
func (pp *ToSendPoolContractTx) Init() {
	heap.Init((*PriorityQueue)(pp))
}
func (pp *ToSendPoolContractTx) AddTransactions(txs []AnyTransaction) {
	for _, tx := range txs {
		heap.Push((*PriorityQueue)(pp), tx.(*ContractChainTransaction))
	}
}
func (pp *ToSendPoolContractTx) GetTransactionsFromPool(numberTx int) []AnyTransaction {
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
