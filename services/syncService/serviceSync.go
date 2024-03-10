package syncServices

import (
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/message"
	"github.com/quad/quad-node/services"
	"github.com/quad/quad-node/tcpip"
	"log"
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
		Chain:   255,
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
	nb := n.GetBytes()
	return nb
}

func Send(addr string, nb []byte) {
	bip := []byte(addr)
	lip := common.GetByteInt16(int16(len(bip)))
	lip = append(lip, bip...)
	nb = append(lip, nb...)
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
	for i := 0; i < 5; i++ {
		go tcpip.StartNewListener(services.SendChanSync, tcpip.SyncTopic[i])
	}
}

func StartSubscribingSyncMsg(ip string, chain uint8) {
	recvChan := make(chan []byte)

	go tcpip.StartNewConnection(ip, recvChan, tcpip.SyncTopic[chain])
	log.Println("Enter connection receiving loop (sync msg)", ip)
Q:

	for {
		select {
		case s := <-recvChan:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
			if len(s) > 2 {
				l := common.GetInt16FromByte(s[:2])
				if len(s) > 2+int(l) {
					ipr := string(s[2 : 2+l])

					OnMessage(ipr, s[2+l:])
				}
			}

		case <-tcpip.Quit:
			break Q
		default:
		}

	}
	log.Println("Exit connection receiving loop (sync msg)", ip)
}
