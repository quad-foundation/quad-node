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

var waitChan chan []byte

func init() {
	waitChan = make(chan []byte)
}

func StartNewListener(sendChan <-chan []byte, topic string) {

	conn, err := Listen("0.0.0.0", ports[topic])
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	defer func() {
		for _, tcpConn := range tcpConnections[topic] {
			tcpConn.Close()
		}
	}()

	go LoopSend(sendChan, topic)
Q:
	for {
		_, err := Accept(topic, conn)
		if err != nil {
			log.Println(err)
			continue
		}
		select {
		case <-Quit:
			break Q
		default:
		}
	}
}

func LoopSend(sendChan <-chan []byte, topic string) {

QUITFOR:
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
					//kip := strings.Split(k, ":")
					if k != MyIP {
						log.Println("send to ipr", k)
						Send(tcpConn0, s[2+l:])
					}
				}

			} else {
				peersMutex.RLock()
				tcpConns := tcpConnections[topic]
				peersMutex.RUnlock()
				//ipr = fmt.Sprintf("%v:%v", ipr, ports[topic])
				tcpConn, ok := tcpConns[ipr]
				if ok {
					log.Println("send to ip", ipr)
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
			break QUITFOR
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

CLOSECONN:
	for {
		r := Receive(topic, tcpConn)
		if r == nil {
			continue
		}
		if len(r) == 7 && string(r) == "QUITFOR" {
			break CLOSECONN
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
		log.Println("receive from ip", ip, topic)
		for _, e := range rs[:lastBytesNum] {
			if len(e) > 0 {
				receiveChan <- append(bmyip, e...)
			}
		}
		select {
		case <-Quit:
			break CLOSECONN
		default:
		}
	}
	receiveChan <- []byte("EXIT")
	CloseAndRemoveConnection(tcpConn)
	fmt.Println("Closing connection (receive)", ip)
}

func CloseAndRemoveConnection(tcpConn *net.TCPConn) {
	if tcpConn == nil {
		return
	}
	peersMutex.Lock()
	for topic, c := range tcpConnections {
		for k, v := range c {
			if tcpConn.RemoteAddr().String() == v.RemoteAddr().String() {
				fmt.Println("Closing connection (send)", topic, k)
				tcpConnections[topic][k].Close()
				delete(tcpConnections[topic], k)
				//ip := strings.Split(k, ":")
				delete(peersConnected, topic+k)
			}
		}
	}
	peersMutex.Unlock()
}
