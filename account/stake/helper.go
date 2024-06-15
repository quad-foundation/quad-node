package stake

import "sync"

var StakingRWMutex sync.RWMutex

func ExtractKeysOfList(m map[int64][]StakingDetail) []int64 {
	keys := []int64{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func ContainsKeyInt64(keys []int64, searchKey int64) bool {
	for _, key := range keys {
		if key == searchKey {
			return true
		}
	}
	return false
}
