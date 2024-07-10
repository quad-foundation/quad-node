package account

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	"math"
	"sync"
)

func Int64toFloat64(value int64) float64 {
	return float64(value) * math.Pow10(-int(common.Decimals))
}

func Int64toFloat64ByDecimals(value int64, decimals uint8) float64 {
	return float64(value) * math.Pow10(-int(decimals))
}

//	func IsDelegatedAccount(ab [common.AddressLength]byte) bool {
//		n := common.GetInt16FromByte(ab[:])
//		if n >= 256 || n == 0 {
//			return false
//		}
//		da := common.GetDelegatedAccountAddress(n)
//		return bytes.Equal(da.GetBytes(), ab[:])
//	}

//
//func IsDEXAccountFromAddress(a common.Address) bool {
//	n := common.GetInt16FromByte(a.GetByte())
//	if n != 256 {
//		return false
//	}
//	da := common.GetDelegatedAccountByteForDEX(n, a.GetByte()[2:])
//	return bytes.Equal(da.GetByte(), a.GetByte())
//}

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

func IntDelegatedAccountFromAddress(a common.Address) (int, error) {
	n := binary.BigEndian.Uint16(a.GetBytes())
	if n < 1 {
		return -1, fmt.Errorf("this is not correct delegated account")
	}
	for _, b := range a.GetBytes()[2:] {
		if b != 0 {
			return -1, fmt.Errorf("this is not correct delegated account")
		}
	}
	da := common.GetDelegatedAccountAddress(int16(n))
	if bytes.Equal(da.GetBytes(), a.GetBytes()) {
		return int(n), nil
	}
	return -1, fmt.Errorf("wrongly formated delegated account")
}
