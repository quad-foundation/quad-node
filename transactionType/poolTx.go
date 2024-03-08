package transactionType

import (
	"bytes"
	"container/heap"
	"github.com/quad/quad-node/common"
	"sync"
)

type Item struct {
	Transaction
	value    [common.HashLength]byte
	priority int64
	index    int
}

func NewItem(tx Transaction, priority int64) *Item {
	hash := [common.HashLength]byte{}
	calcHash := tx.GetHash()
	//if err != nil {
	//	return nil
	//}
	copy(hash[:], calcHash.GetBytes())
	return &Item{
		Transaction: tx,
		value:       hash,
		priority:    priority,
	}
}

type TransactionPool struct {
	transactions    map[[common.HashLength]byte]Transaction
	priorityQueue   PriorityQueue
	maxTransactions int
	rwmutex         sync.RWMutex
}

func NewTransactionPool(maxTransactions int) *TransactionPool {
	return &TransactionPool{
		transactions:    make(map[[common.HashLength]byte]Transaction),
		priorityQueue:   make(PriorityQueue, 0),
		maxTransactions: maxTransactions,
	}
}
func (tp *TransactionPool) AddTransaction(tx Transaction) {
	var hash [common.HashLength]byte
	copy(hash[:], tx.GetHash().GetBytes())
	tp.rwmutex.Lock()
	defer tp.rwmutex.Unlock()
	if _, exists := tp.transactions[hash]; !exists {
		tp.transactions[hash] = tx
		item := NewItem(tx, tx.GetGasPrice())
		heap.Push(&tp.priorityQueue, item)
		if tp.priorityQueue.Len() > tp.maxTransactions {
			removed := heap.Pop(&tp.priorityQueue).(*Item)
			delete(tp.transactions, removed.value)
		}
	}
}
func (tp *TransactionPool) PeekTransactions(n int) []Transaction {
	if n > len(tp.transactions) {
		n = len(tp.transactions)
	}
	hash := [common.HashLength]byte{}
	topTransactions := []Transaction{}
	tp.rwmutex.RLock()
	defer tp.rwmutex.RUnlock()
	for i := 0; i < n; i++ {
		if len(tp.priorityQueue) > i {
			transaction := *tp.priorityQueue[i]
			copy(hash[:], transaction.GetHash().GetBytes())
			topTransactions = append(topTransactions, tp.transactions[hash])
		}
	}
	return topTransactions
}
func (tp *TransactionPool) RemoveTransactionByHash(hash []byte) {
	h := [common.HashLength]byte{}
	copy(h[:], hash)
	tp.rwmutex.Lock()
	defer tp.rwmutex.Unlock()
	if _, exists := tp.transactions[h]; exists {
		for i := 0; i < tp.priorityQueue.Len(); i++ {
			h2 := (*tp.priorityQueue[i]).GetHash().GetBytes()
			if bytes.Equal(h2, h[:]) {
				heap.Remove(&tp.priorityQueue, i)
				break
			}
		}
		delete(tp.transactions, h)
	}
}
func (tp *TransactionPool) PopTransactionByHash(hash []byte) Transaction {
	h := [common.HashLength]byte{}
	copy(h[:], hash)
	tp.rwmutex.Lock()
	defer tp.rwmutex.Unlock()
	if _, exists := tp.transactions[h]; exists {
		for i := 0; i < tp.priorityQueue.Len(); i++ {
			h2 := (*tp.priorityQueue[i]).GetHash().GetBytes()
			if bytes.Equal(h2, h[:]) {
				tx := tp.transactions[h]
				heap.Remove(&tp.priorityQueue, i)
				delete(tp.transactions, h)
				return tx
			}
		}
	}
	return EmptyTransaction()
}
func (tp *TransactionPool) NumberOfTransactions() int {
	tp.rwmutex.RLock()
	defer tp.rwmutex.RUnlock()
	return len(tp.transactions)
}
