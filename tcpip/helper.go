package tcpip

import (
	"github.com/quad-foundation/quad-node/common"
	"log"
	"sync"
)

const bannedTime int64 = 10 //1440 * 7 * 6 // 7 days
var bannedIP map[string]int64
var bannedIPMutex sync.RWMutex

func init() {
	bannedIP = map[string]int64{}
}
func IsIPBanned(ip string, h int64, topic [2]byte) bool {
	bannedIPMutex.RLock()
	defer bannedIPMutex.RUnlock()
	ip = string(topic[:]) + ip
	if hbanned, ok := bannedIP[ip]; ok {
		if h < hbanned {
			return true
		}
	}
	return false
}

func BanIP(ip string, topic [2]byte) {
	bannedIPMutex.Lock()
	defer bannedIPMutex.Unlock()
	log.Println("banning ", ip, " with topic ", topic[:])
	ip = string(topic[:]) + ip
	bannedIP[ip] = common.GetHeight() + bannedTime

	PeersMutex.RLock()
	tcpConns := tcpConnections[topic]
	tcpConn, ok := tcpConns[ip]
	PeersMutex.RUnlock()
	if ok {
		CloseAndRemoveConnection(tcpConn)
	}

}
