package common

import (
	"sync"
	"sync/atomic"
)

var height int64
var heightMax int64
var heightMutex sync.RWMutex
var heightMaxMutex sync.RWMutex
var BlockMutex sync.RWMutex
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

func GetHeightMax() int64 {
	heightMaxMutex.RLock()
	defer heightMaxMutex.RUnlock()
	return heightMax
}

func SetHeightMax(hmax int64) {
	heightMaxMutex.Lock()
	defer heightMaxMutex.Unlock()
	heightMax = hmax
}
