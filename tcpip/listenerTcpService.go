package tcpip

import (
	"bytes"
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	"log"
	"net"
	"strings"
	"time"
)

func StartNewListener(sendChan <-chan []byte, topic string) {

	conn, err := Listen("0.0.0.0", ports[topic])
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	defer func() {
		peersMutex.Lock()
		defer peersMutex.Unlock()
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
func LoopSend(sendChan <-chan []byte, topic string) {
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
				peersMutex.RLock()
				tmpConn := tcpConnections[topic]
				peersMutex.RUnlock()
				for k, tcpConn0 := range tmpConn {
					if k != MyIP {
						//log.Println("send to ipr", k)
						Send(tcpConn0, s[2+l:])
					}
				}
			} else {
				peersMutex.RLock()
				tcpConns := tcpConnections[topic]
				peersMutex.RUnlock()
				tcpConn, ok := tcpConns[ipr]
				if ok {
					//log.Println("send to ip", ipr)
					Send(tcpConn, s[2+l:])
				} else {
					fmt.Println("no connection to given ip", ipr, topic)
				}
			}
		case b := <-waitChan:
			s := string(b)
			if s == topic {
				time.Sleep(time.Millisecond * 10)
			}
		case <-Quit:
			return
		default:
		}
	}
}
func StartNewConnection(ip string, receiveChan chan []byte, topic string) {
	var tcpConn *net.TCPConn
	a := strings.Split(ip, ":")
	ip = a[0]
	ipport := fmt.Sprint(ip, ":", ports[topic])
	if ip == "127.0.0.1" {
		ipport = fmt.Sprint(":", ports[topic])
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
	raddr := tcpConn.RemoteAddr().String()
	fmt.Println("New connection from address", raddr, topic)
	lastBytes := []byte{}
	lastBytesNum := 0
	for {
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
			if len(r) == 7 && string(r) == "QUITFOR" {
				receiveChan <- []byte("EXIT")
				CloseAndRemoveConnection(tcpConn)
				fmt.Println("Closing connection (receive)", ip)
				return
			}
			if len(r) == 4 && string(r) == "WAIT" {
				waitChan <- []byte(topic)
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
	peersMutex.Lock()
	defer peersMutex.Unlock()
	for topic, c := range tcpConnections {
		for k, v := range c {
			if tcpConn.RemoteAddr().String() == v.RemoteAddr().String() {
				fmt.Println("Closing connection (send)", topic, k)
				tcpConnections[topic][k].Close()
				delete(tcpConnections[topic], k)
				delete(peersConnected, topic+k)
			}
		}
	}
}
