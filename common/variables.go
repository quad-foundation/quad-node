package common

import (
	"sync"
	"sync/atomic"
)

var height int64
var HeightMutex sync.RWMutex
var BalanceMutex sync.RWMutex
var BlockMutex sync.Mutex
var SyncingMutex sync.Mutex
var IsSyncing = atomic.Bool{}

func GetHeight() int64 {
	HeightMutex.RLock()
	h := height
	HeightMutex.RUnlock()
	return h
}

func SetHeight(h int64) {
	HeightMutex.Lock()
	height = h
	HeightMutex.Unlock()
}

func CheckHeight(chain uint8, heightToCheck int64) bool {
	return GetChainForHeight(heightToCheck) == chain
}

func GetChainForHeight(heightToCheck int64) uint8 {
	chainProper := heightToCheck % 5
	return uint8(chainProper)
}
