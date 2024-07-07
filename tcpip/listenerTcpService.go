package tcpip

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	"log"
	"net"
	"strings"
	"time"
)

func StartNewListener(sendChan <-chan []byte, topic [2]byte) {

	conn, err := Listen("0.0.0.0", Ports[topic])
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
	for {
		select {
		case s := <-sendChan:
			l := common.GetInt16FromByte(s[:2])
			if len(s) < int(l)+2 {
				log.Println("wrong message send")
				continue
			}
			ipr := string(s[2 : 2+l])
			if ipr == "0.0.0.0" {
				PeersMutex.RLock()
				tmpConn := tcpConnections[topic]
				PeersMutex.RUnlock()
				for k, tcpConn0 := range tmpConn {
					if k != MyIP {
						//log.Println("send to ipr", k)
						Send(tcpConn0, s[2+l:])
					}
				}
			} else {
				PeersMutex.RLock()
				tcpConns := tcpConnections[topic]
				tcpConn, ok := tcpConns[ipr]
				PeersMutex.RUnlock()
				if ok {
					//log.Println("send to ip", ipr)
					Send(tcpConn, s[2+l:])
				} else {
					fmt.Println("no connection to given ip", ipr, topic)
					BanIP(ipr, topic)
				}
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
func StartNewConnection(ip string, receiveChan chan []byte, topic [2]byte) {
	var tcpConn *net.TCPConn
	a := strings.Split(ip, ":")
	ip = a[0]
	ipport := fmt.Sprint(ip, ":", Ports[topic])
	if ip == "127.0.0.1" {
		ipport = fmt.Sprint(":", Ports[topic])
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", ipport)
	if err != nil {
		log.Println("cannot create tcp address", err)
		return
	}
	bmyip := []byte(ip)
	lmyip := int16(len(bmyip))
	blmyip := common.GetByteInt16(lmyip)
	bmyip = append(blmyip, bmyip...)
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
			if string(r) == "<-CLS->" {
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

			if len(r) == 7 && string(r) == "QUITFOR" {
				receiveChan <- []byte("EXIT")
				CloseAndRemoveConnection(tcpConn)
				fmt.Println("Closing connection (receive)", ip)
				return
			}
			if len(r) == 4 && string(r) == "WAIT" {
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
					receiveChan <- append(bmyip, e...)
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
	for topic, c := range tcpConnections {
		for k, v := range c {
			if tcpConn.RemoteAddr().String() == v.RemoteAddr().String() {
				fmt.Println("Closing connection (send)", topic, k)
				tcpConnections[topic][k].Close()
				delete(tcpConnections[topic], k)
				delete(peersConnected, string(topic[:])+k)
			}
		}
	}
}
