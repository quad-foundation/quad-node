package common

import (
	"sync"
	"sync/atomic"
)

var height int64
var heightMutex sync.RWMutex
var BalanceMutex sync.RWMutex
var BlockMutex sync.RWMutex
var SyncingMutex sync.Mutex
var IsSyncing = atomic.Bool{}

func GetHeight() int64 {
	heightMutex.RLock()
	defer heightMutex.RUnlock()
	return height
}

func SetHeight(h int64) {
	heightMutex.Lock()
	defer heightMutex.Unlock()
	height = h
}
