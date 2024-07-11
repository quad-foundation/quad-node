package tcpip

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

func StartNewListener(sendChan <-chan []byte, topic [2]byte) {

	conn, err := Listen([4]byte{0, 0, 0, 0}, Ports[topic])
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	defer func() {
		PeersMutex.Lock()
		defer PeersMutex.Unlock()
		for _, tcpConn := range tcpConnections[topic] {
			tcpConn.Close()
		}
	}()
	go LoopSend(sendChan, topic)
	for {
		select {
		case <-Quit:
			return
		default:
			_, err := Accept(topic, conn)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

//func worker(sendChan <-chan []byte, topic [2]byte, wg *sync.WaitGroup) {
//	defer wg.Done()
//	var ipr [4]byte
//	for s := range sendChan {
//		if len(s) > 4 {
//			copy(ipr[:], s[:4])
//		} else {
//			log.Println("wrong message")
//			continue
//		}
//		PeersMutex.RLock()
//		if bytes.Equal(ipr[:], []byte{0, 0, 0, 0}) {
//			tmpConn := tcpConnections[topic]
//			for k, tcpConn0 := range tmpConn {
//				if !bytes.Equal(k[:], MyIP[:]) {
//					Send(tcpConn0, s[4:])
//				}
//			}
//		} else {
//			tcpConns := tcpConnections[topic]
//			tcpConn, ok := tcpConns[ipr]
//			if ok {
//				Send(tcpConn, s[4:])
//			} else {
//				// Handle no connection case
//			}
//		}
//		PeersMutex.RUnlock()
//	}
//}
//func LoopSend(sendChan <-chan []byte, topic [2]byte, numWorkers int) {
//	var wg sync.WaitGroup
//	// Start worker goroutines
//	for i := 0; i < numWorkers; i++ {
//		wg.Add(1)
//		go worker(sendChan, topic, &wg)
//	}
//	for {
//		select {
//		case b := <-waitChan:
//			if bytes.Equal(b, topic[:]) {
//				time.Sleep(time.Millisecond * 10)
//			}
//		case <-Quit:
//			//close(sendChan)
//			wg.Wait()
//			return
//		default:
//		}
//	}
//}

func LoopSend(sendChan <-chan []byte, topic [2]byte) {
	var ipr [4]byte
	for {
		select {
		case s := <-sendChan:
			if len(s) > 4 {
				copy(ipr[:], s[:4])
			} else {
				log.Println("wrong message")
				continue
			}
			PeersMutex.RLock()
			if bytes.Equal(ipr[:], []byte{0, 0, 0, 0}) {

				tmpConn := tcpConnections[topic]
				for k, tcpConn0 := range tmpConn {
					if !bytes.Equal(k[:], MyIP[:]) {
						//log.Println("send to ipr", k)
						Send(tcpConn0, s[4:])
					}
				}
			} else {
				tcpConns := tcpConnections[topic]
				tcpConn, ok := tcpConns[ipr]
				if ok {
					//log.Println("send to ip", ipr)
					Send(tcpConn, s[4:])
				} else {
					//fmt.Println("no connection to given ip", ipr, topic)
					//BanIP(ipr, topic)
				}

			}
			PeersMutex.RUnlock()
		case b := <-waitChan:
			if bytes.Equal(b, topic[:]) {
				time.Sleep(time.Millisecond * 10)
			}
		case <-Quit:
			return
		default:
		}
	}
}

func StartNewConnection(ip [4]byte, receiveChan chan []byte, topic [2]byte) {
	ipport := fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], Ports[topic])
	if bytes.Equal(ip[:], []byte{127, 0, 0, 1}) {
		ipport = fmt.Sprintf(":%d", Ports[topic])
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", ipport)
	if err != nil {
		log.Println("cannot create tcp address", err)
		return
	}
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Println("connection to ip was unsuccessful", ip, topic, err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			PeersMutex.Lock()
			defer PeersMutex.Unlock()
			tcpConn.Close()
			log.Println("recover (receive Msg)", r)
		}
	}()
	fmt.Println("New connection from address", tcpConn.RemoteAddr().String(), topic)
	lastBytes := []byte{}
	reconnectionTries := 0
	resetNumber := 0
	for {
		resetNumber++
		if resetNumber%100 == 0 {
			reconnectionTries = 0
		}
		select {
		case <-Quit:
			receiveChan <- []byte("EXIT")
			CloseAndRemoveConnection(tcpConn)
			fmt.Println("Closing connection (receive)", ip)
			return
		default:
			r := Receive(topic, tcpConn)
			if r == nil {
				continue
			}
			if bytes.Equal(r, []byte("<-CLS->")) {
				if reconnectionTries > 50 {
					CloseAndRemoveConnection(tcpConn)
					fmt.Println("Closing connection (receive)", ip)
					return
				}
				reconnectionTries++
				tcpConn, err = net.DialTCP("tcp", nil, tcpAddr)
				if err != nil {
					log.Println("connection to ip was unsuccessful", ip, topic, err)
				}
				continue
			}
			if bytes.Equal(r, []byte("QUITFOR")) {
				receiveChan <- []byte("EXIT")
				CloseAndRemoveConnection(tcpConn)
				fmt.Println("Closing connection (receive)", ip)
				return
			}
			if bytes.Equal(r, []byte("WAIT")) {
				waitChan <- topic[:]
				continue
			}
			r = append(lastBytes, r...)
			rs := bytes.Split(r, []byte("<-END->"))
			if !bytes.Equal(r[len(r)-7:], []byte("<-END->")) {
				lastBytes = rs[len(rs)-1]
			} else {
				lastBytes = []byte{}
			}
			for _, e := range rs[:len(rs)-1] {
				if len(e) > 0 {
					receiveChan <- append(ip[:], e...)
				}
			}
		}
	}
}

func CloseAndRemoveConnection(tcpConn *net.TCPConn) {
	if tcpConn == nil {
		return
	}
	PeersMutex.RLock()
	defer PeersMutex.RUnlock()
	topicipBytes := [6]byte{}
	for topic, c := range tcpConnections {
		for k, v := range c {
			if tcpConn.RemoteAddr().String() == v.RemoteAddr().String() {
				PeersMutex.RUnlock()
				PeersMutex.Lock()
				fmt.Println("Closing connection (send)", topic, k)
				tcpConnections[topic][k].Close()
				copy(topicipBytes[:], append(topic[:], k[:]...))
				delete(tcpConnections[topic], k)
				delete(peersConnected, topicipBytes)
				PeersMutex.Unlock()
				PeersMutex.RLock()
			}
		}
	}
}
