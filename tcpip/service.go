package tcpip

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var peersConnected = map[string]string{}
var peersMutex sync.RWMutex
var oldPeers = map[string]string{}
var tcpConnections = map[string]map[string]*net.TCPConn{}
var Quit chan os.Signal
var CheckCount int64

var (
	TransactionTopic = [5]string{"T0", "T1", "T2", "T3", "T4"}
	NonceTopic       = [5]string{"N0", "N1", "N2", "N3", "N4"}
	SelfNonceTopic   = [5]string{"S0", "S1", "S2", "S3", "S4"}
	SyncTopic        = [5]string{"B0", "B1", "B2", "B3", "B4"}
)

var ports = map[string]int{
	TransactionTopic[0]: 9091,
	TransactionTopic[1]: 9092,
	TransactionTopic[2]: 9093,
	TransactionTopic[3]: 9094,
	TransactionTopic[4]: 9095,
	NonceTopic[0]:       8091,
	NonceTopic[1]:       8092,
	NonceTopic[2]:       8093,
	NonceTopic[3]:       8094,
	NonceTopic[4]:       8095,
	SelfNonceTopic[0]:   7091,
	SelfNonceTopic[1]:   7092,
	SelfNonceTopic[2]:   7093,
	SelfNonceTopic[3]:   7094,
	SelfNonceTopic[4]:   7095,
	SyncTopic[0]:        6091,
	SyncTopic[1]:        6092,
	SyncTopic[2]:        6093,
	SyncTopic[3]:        6094,
	SyncTopic[4]:        6095,
}

var MyIP string

func init() {
	Quit = make(chan os.Signal)
	signal.Notify(Quit, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	// rather here should be external IP which should be set in .env file
	MyIP = getInternalIp()

	for k, _ := range ports {
		tcpConnections[k] = map[string]*net.TCPConn{}
	}
}

type IP struct {
	Query string
}

// getInternalIp function to retrieve intranet IP of node
func getInternalIp() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println("Can not obtain net interface")
		return ""
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Println("Can not get net addresses")
			return ""
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.IsLoopback() {
				continue
			}
			if ip.IsPrivate() {
				return ip.String()
			}
		}
	}
	return ""
}

// Listen internal function which set server for ip and port
func Listen(ip string, port int) (*net.TCPListener, error) {
	ipport := fmt.Sprint(ip, ":", port)
	protocol := "tcp"

	//Build the address
	addr, err := net.ResolveTCPAddr(protocol, ipport)
	if err != nil {
		log.Println("Wrong Address", err)
		return nil, err
	}
	conn, err := net.ListenTCP(protocol, addr)
	if err != nil {
		log.Printf("Some error %v\n", err)
		return nil, err
	}

	return conn, nil
}

func Accept(topic string, conn *net.TCPListener) (*net.TCPConn, error) {
	tcpConn, err := conn.AcceptTCP()
	if err == nil {
		//tcpConn.SetWriteBuffer(1024 * 1024)
		NewConnectionPeer(topic, tcpConn)
		return tcpConn, nil
	} else {
		return nil, fmt.Errorf("no connection available yet")
	}
}

func NewConnectionPeer(topic string, tcpConn *net.TCPConn) {

	raddr := tcpConn.RemoteAddr().String()
	ra := strings.Split(raddr, ":")
	addrRemote := ra[0]
	topicip := topic + ra[0]
	peersMutex.Lock()
	if t, ok := peersConnected[topicip]; !ok || t != topic {
		log.Println("New connection from address", addrRemote, topic)
		tcpConnections[topic][addrRemote] = tcpConn
		peersConnected[topicip] = topic
	}
	peersMutex.Unlock()
}

func Send(conn *net.TCPConn, message []byte) {

	message = append(message, []byte("<-END->")...)
	n, err := conn.Write(message)
	if err != nil {
		log.Printf("Cann't send response %v", err)
		if err == io.EOF {
			log.Println("buffer is full (send)")
			//CloseAndRemoveConnection(conn)
			time.Sleep(time.Millisecond * 10)
		}
		if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNABORTED) {
			CloseAndRemoveConnection(conn)
		}
	}
	if n != len(message) && n > 0 {
		log.Println("Not whole message was send")
	}
}

func Receive(topic string, conn *net.TCPConn) []byte {

	buf := make([]byte, 1024*1024) // big buffer 1 MB - full block

	n, err := conn.Read(buf[:])
	if err != nil {
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			return nil
		}

		if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNABORTED) {
			log.Print("This is broken pipe error")
			return []byte("QUITFOR")
		}
		if err == io.EOF {
			//log.Println("buffer is full (receive)")
			return []byte("WAIT")
		}
		return nil
	}
	if CheckCount > 10 {
		NewConnectionPeer(topic, conn)
		CheckCount = 0
	}
	CheckCount++
	return buf[:n]
}

func LookUpForNewPeersToConnect(chanPeer chan string) {
	for {
		peersMutex.RLock()
		for topicip, topic := range peersConnected {
			_, ok := oldPeers[topicip]
			if ok == false {
				log.Println("Found new peer with ip", topicip)
				oldPeers[topicip] = topic
				chanPeer <- topicip
			}
		}
		for topicip, _ := range oldPeers {
			_, ok := peersConnected[topicip]
			if ok == false {
				log.Println("New peer is deleted with ip", topicip)
				delete(oldPeers, topicip)
			}
		}
		peersMutex.RUnlock()
		time.Sleep(time.Second)
	}
}
