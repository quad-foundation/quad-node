package syncServices

import (
	"github.com/quad-foundation/quad-node/blocks"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/message"
	"github.com/quad-foundation/quad-node/services"
	"github.com/quad-foundation/quad-node/tcpip"
	"log"
	"net"
	"time"
)

func InitSyncService() {
	services.SendMutexSync.Lock()
	services.SendChanSync = make(chan []byte)

	services.SendMutexSync.Unlock()
	startPublishingSyncMsg()
	time.Sleep(time.Second)
	go sendSyncMsgInLoop()
}

func generateSyncMsgHeight() []byte {
	h := common.GetHeight()
	bm := message.BaseMessage{
		Head:    []byte("hi"),
		ChainID: common.GetChainID(),
	}
	n := message.TransactionsMessage{
		BaseMessage:       bm,
		TransactionsBytes: map[[2]byte][][]byte{},
	}
	n.TransactionsBytes[[2]byte{'L', 'H'}] = [][]byte{common.GetByteInt64(h)}
	lastBlockHash, err := blocks.LoadHashOfBlock(h)
	if err != nil {
		log.Println("Can not obtain root hashes from DB", err)
		return []byte("")
	}
	n.TransactionsBytes[[2]byte{'L', 'B'}] = [][]byte{lastBlockHash}
	tcpip.PeersMutex.RLock()
	peers := tcpip.GetIPsConnected()
	tcpip.PeersMutex.RUnlock()
	n.TransactionsBytes[[2]byte{'P', 'P'}] = peers
	nb := n.GetBytes()
	return nb
}

func generateSyncMsgGetHeaders(height int64) []byte {
	if height <= 0 {
		return nil
	}
	eHeight := height
	h := common.GetHeight()
	bHeight := height - common.NumberOfHashesInBucket
	if bHeight <= 0 {
		bHeight = 0
	}
	if bHeight > h {
		bHeight = h
		eHeight = h + common.NumberOfHashesInBucket
		if eHeight > height {
			eHeight = height
		}
	}
	bm := message.BaseMessage{
		Head:    []byte("gh"),
		ChainID: common.GetChainID(),
	}
	n := message.TransactionsMessage{
		BaseMessage:       bm,
		TransactionsBytes: map[[2]byte][][]byte{},
	}
	n.TransactionsBytes[[2]byte{'B', 'H'}] = [][]byte{common.GetByteInt64(bHeight)}
	n.TransactionsBytes[[2]byte{'E', 'H'}] = [][]byte{common.GetByteInt64(eHeight)}
	nb := n.GetBytes()
	return nb
}

func generateSyncMsgSendHeaders(bHeight int64, height int64) []byte {
	if height < 0 {
		log.Println("height cannot be smaller than 0")
		return []byte{}
	}
	h := common.GetHeight()
	if height > h {
		log.Println("Warning: height cannot be larger than last height")
		height = h
	}
	if bHeight < 0 || bHeight > height {
		log.Println("starting height cannot be smaller than 0")
		return []byte{}
	}
	bm := message.BaseMessage{
		Head:    []byte("sh"),
		ChainID: common.GetChainID(),
	}
	n := message.TransactionsMessage{
		BaseMessage:       bm,
		TransactionsBytes: map[[2]byte][][]byte{},
	}
	indices := [][]byte{}
	blcks := [][]byte{}
	for i := bHeight; i <= height; i++ {
		indices = append(indices, common.GetByteInt64(i))
		block, err := blocks.LoadBlock(i)
		if err != nil {
			log.Println(err)
			return []byte{}
		}
		blcks = append(blcks, block.GetBytes())
	}
	n.TransactionsBytes[[2]byte{'I', 'H'}] = indices
	n.TransactionsBytes[[2]byte{'H', 'V'}] = blcks
	nb := n.GetBytes()
	return nb
}

func SendHeaders(addr string, bHeight int64, height int64) {
	n := generateSyncMsgSendHeaders(bHeight, height)
	Send(addr, n)
}

func SendGetHeaders(addr string, height int64) {
	n := generateSyncMsgGetHeaders(height)
	Send(addr, n)
}

func Send(addr string, nb []byte) {
	nb = append(net.ParseIP(addr).To4(), nb...)
	services.SendMutexSync.Lock()
	services.SendChanSync <- nb
	services.SendMutexSync.Unlock()
}

func sendSyncMsgInLoop() {
	for range time.Tick(time.Second * 5) {
		n := generateSyncMsgHeight()
		Send("0.0.0.0", n)
	}
}

func startPublishingSyncMsg() {

	go tcpip.StartNewListener(services.SendChanSync, tcpip.SyncTopic)
}

func StartSubscribingSyncMsg(ip string) {
	recvChan := make(chan []byte, 100) // Use a buffered channel
	quit := false
	go tcpip.StartNewConnection(ip, recvChan, tcpip.SyncTopic)
	log.Println("Enter connection receiving loop (sync msg)", ip)
	for !quit {
		select {
		case s := <-recvChan:
			if len(s) == 4 && string(s) == "EXIT" {
				quit = true
				break
			}
			if len(s) > 4 {
				ipr := net.IPv4(s[0], s[1], s[2], s[3]).String()
				OnMessage(ipr, s[4:])
			}
		case <-tcpip.Quit:
			quit = true
		default:
			// Optional: Add a small sleep to prevent busy-waiting
			time.Sleep(time.Millisecond)
		}
	}
	log.Println("Exit connection receiving loop (sync msg)", ip)
}

//func StartSubscribingSyncMsg(ip string) {
//	recvChan := make(chan []byte)
//
//	go tcpip.StartNewConnection(ip, recvChan, tcpip.SyncTopic)
//	log.Println("Enter connection receiving loop (sync msg)", ip)
//Q:
//
//	for {
//		select {
//		case s := <-recvChan:
//			if len(s) == 4 && string(s) == "EXIT" {
//				break Q
//			}
//			if len(s) > 4 {
//				ipr := net.IPv4(s[0], s[1], s[2], s[3]).String()
//				OnMessage(ipr, s[4:])
//			}
//
//		case <-tcpip.Quit:
//			break Q
//		default:
//		}
//
//	}
//	log.Println("Exit connection receiving loop (sync msg)", ip)
//}
