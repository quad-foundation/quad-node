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
			if bytes.Compare(ipr[:], []byte{0, 0, 0, 0}) == 0 {
				PeersMutex.RLock()
				tmpConn := tcpConnections[topic]
				for k, tcpConn0 := range tmpConn {
					if bytes.Compare(k[:], MyIP[:]) != 0 {
						//log.Println("send to ipr", k)
						Send(tcpConn0, s[4:])
					}
				}
				PeersMutex.RUnlock()
			} else {
				PeersMutex.RLock()
				tcpConns := tcpConnections[topic]
				tcpConn, ok := tcpConns[ipr]
				if ok {
					//log.Println("send to ip", ipr)
					Send(tcpConn, s[4:])
				} else {
					//fmt.Println("no connection to given ip", ipr, topic)
					//BanIP(ipr, topic)
				}
				PeersMutex.RUnlock()
			}
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
	var tcpConn *net.TCPConn

	ipport := fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], Ports[topic])
	if bytes.Compare(ip[:], []byte{127, 0, 0, 1}) == 0 {
		ipport = fmt.Sprint(":", Ports[topic])
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", ipport)
	if err != nil {
		log.Println("cannot create tcp address", err)
		return
	}

	tcpConn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("connection to ip was unsuccessful", ip, topic, err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			PeersMutex.Lock()
			defer PeersMutex.Unlock()
			tcpConn.Close()
			log.Println("recover (receive Msg)", r)
		}
		return
	}()
	raddr := tcpConn.RemoteAddr().String()
	fmt.Println("New connection from address", raddr, topic)
	lastBytes := []byte{}
	lastBytesNum := 0
	reconectionTries := 0
	resetNumber := 0

	for {
		resetNumber++
		if resetNumber%100 == 0 {
			reconectionTries = 0
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
			if bytes.Compare(r, []byte("<-CLS->")) == 0 {
				if reconectionTries > 50 {
					CloseAndRemoveConnection(tcpConn)
					fmt.Println("Closing connection (receive)", ip)
					return
				} else {
					reconectionTries++
				}
				tcpConn, err = net.DialTCP("tcp", nil, tcpAddr)
				if err != nil {
					fmt.Println("connection to ip was unsuccessful", ip, topic, err)
				}
				continue
			}

			if len(r) == 7 && bytes.Compare(r, []byte("QUITFOR")) == 0 {
				receiveChan <- []byte("EXIT")
				CloseAndRemoveConnection(tcpConn)
				fmt.Println("Closing connection (receive)", ip)
				return
			}
			if len(r) == 4 && bytes.Compare(r, []byte("WAIT")) == 0 {
				waitChan <- topic[:]
				continue
			}
			r = append(lastBytes, r...)
			rs := bytes.Split(r, []byte("<-END->"))
			if bytes.Compare(r[len(r)-7:], []byte("<-END->")) != 0 {
				lastBytes = rs[len(rs)-1]
				lastBytesNum = len(rs) - 1
			} else {
				lastBytes = []byte{}
				lastBytesNum = len(rs)
			}
			//log.Println("receive from ip", ip, topic)
			for _, e := range rs[:lastBytesNum] {
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
	PeersMutex.Lock()
	defer PeersMutex.Unlock()
	topicipBytes := [6]byte{}
	for topic, c := range tcpConnections {
		for k, v := range c {
			if tcpConn.RemoteAddr().String() == v.RemoteAddr().String() {
				fmt.Println("Closing connection (send)", topic, k)
				tcpConnections[topic][k].Close()
				copy(topicipBytes[:], append(topic[:], k[:]...))
				delete(tcpConnections[topic], k)
				delete(peersConnected, topicipBytes)
			}
		}
	}
}
