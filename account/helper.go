package account

import (
	"github.com/quad/quad-node/common"
	"math"
)

func Int64toFloat64(value int64) float64 {
	return float64(value) * math.Pow10(-int(common.Decimals))
}

func Int64toFloat64ByDecimals(value int64, decimals uint8) float64 {
	return float64(value) * math.Pow10(-int(decimals))
}

//
//func IsDelegatedAccount(ab [common.AddressLength]byte) bool {
//	n := common.GetInt16FromByte(ab[:])
//	if n >= 256 || n == 0 {
//		return false
//	}
//	da := common.GetDelegatedAccountAddress(n)
//	return bytes.Compare(da.GetBytes(), ab[:]) == 0
//}
//
//func IsDelegatedAccountFromAddress(a common.Address) bool {
//	n := common.GetInt16FromByte(a.GetByte())
//	if !(n > 0 && n < 256) {
//		return false
//	}
//	da := common.GetDelegatedAccountAddress(n)
//	return bytes.Compare(da.GetByte(), a.GetByte()) == 0
//}
//
//func IsDEXAccountFromAddress(a common.Address) bool {
//	n := common.GetInt16FromByte(a.GetByte())
//	if n != 256 {
//		return false
//	}
//	da := common.GetDelegatedAccountByteForDEX(n, a.GetByte()[2:])
//	return bytes.Compare(da.GetByte(), a.GetByte()) == 0
//}
