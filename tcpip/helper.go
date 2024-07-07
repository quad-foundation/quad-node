package tcpip

import (
	"github.com/quad-foundation/quad-node/common"
	"sync"
)

const bannedTime int64 = 1440 * 7 * 6 // 7 days
var bannedIP map[string]int64
var bannedIPMutex sync.RWMutex

func init() {
	bannedIP = map[string]int64{}
}
func IsIPBanned(ip string, h int64) bool {
	bannedIPMutex.RLock()
	defer bannedIPMutex.RUnlock()
	if hbanned, ok := bannedIP[ip]; ok {
		if h < hbanned {
			return true
		}
	}
	return false
}

func BanIP(ip string) {
	bannedIPMutex.Lock()
	defer bannedIPMutex.Unlock()
	bannedIP[ip] = common.GetHeight() + bannedTime
}
