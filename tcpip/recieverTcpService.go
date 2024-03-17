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

var (
	peersConnected   = map[string]string{}
	oldPeers         = map[string]string{}
	CheckCount       = 0
	receiveMutex     = &sync.Mutex{}
	waitChan         = make(chan []byte)
	tcpConnections   = make(map[[2]byte]map[string]*net.TCPConn)
	peersMutex       = &sync.RWMutex{}
	Quit             chan os.Signal
	TransactionTopic = [5][2]byte{{'T', 0}, {'T', 1}, {'T', 2}, {'T', 3}, {'T', 4}}
	NonceTopic       = [5][2]byte{{'N', 0}, {'N', 1}, {'N', 2}, {'N', 3}, {'N', 4}}
	SelfNonceTopic   = [5][2]byte{{'S', 0}, {'S', 1}, {'S', 2}, {'S', 3}, {'S', 4}}
	SyncTopic        = [5][2]byte{{'B', 0}, {'B', 1}, {'B', 2}, {'B', 3}, {'B', 4}}
	RPCTopic         = [2]byte{'R', 'P'}
)
var Ports = map[[2]byte]int{
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
	RPCTopic:            9009,
}
var MyIP string

func init() {
	Quit = make(chan os.Signal)
	signal.Notify(Quit, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	MyIP = getInternalIp()
	for k := range Ports {
		tcpConnections[k] = map[string]*net.TCPConn{}
	}
}

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
func Listen(ip string, port int) (*net.TCPListener, error) {
	ipport := fmt.Sprint(ip, ":", port)
	protocol := "tcp"
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

func Accept(topic [2]byte, conn *net.TCPListener) (*net.TCPConn, error) {
	tcpConn, err := conn.AcceptTCP()
	if err != nil {
		return nil, fmt.Errorf("error accepting connection: %w", err)
	}
	RegisterPeer(topic, tcpConn)
	return tcpConn, nil
}

func Send(conn *net.TCPConn, message []byte) {
	message = append(message, []byte("<-END->")...)
	_, err := conn.Write(message)
	if err != nil {
		log.Printf("Can't send response: %v", err)
		if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNABORTED) {
			CloseAndRemoveConnection(conn)
		}
	}
}

// Receive reads data from the connection and handles errors
func Receive(topic [2]byte, conn *net.TCPConn) []byte {
	const bufSize = 1048576
	buf := make([]byte, bufSize)
	n, err := conn.Read(buf)

	if err != nil {
		handleConnectionError(err, topic, conn)
		return []byte("<-CLS->")
	}

	return buf[:n]
}

// handleConnectionError logs different connection errors and tries to reconnect if necessary
func handleConnectionError(err error, topic [2]byte, conn *net.TCPConn) {
	switch {
	case errors.Is(err, syscall.EPIPE), errors.Is(err, syscall.ECONNRESET), errors.Is(err, syscall.ECONNABORTED):
		log.Print("This is a broken pipe error. Attempting to reconnect...")
	case err == io.EOF:
		log.Print("Connection closed by peer. Attempting to reconnect...")
	default:
		log.Printf("Unexpected error: %v", err)
	}
	// Close the current connection
	conn.Close()
}

// RegisterPeer registers a new peer connection
func RegisterPeer(topic [2]byte, tcpConn *net.TCPConn) {
	raddr := tcpConn.RemoteAddr().String()
	ra := strings.Split(raddr, ":")
	addrRemote := ra[0]
	topicip := string(topic[:]) + addrRemote

	peersMutex.Lock()
	defer peersMutex.Unlock()

	if _, ok := peersConnected[topicip]; !ok {
		log.Println("New connection from address", addrRemote, "on topic", topic)
		// Initialize the map for the topic if it doesn't exist
		if _, ok := tcpConnections[topic]; !ok {
			tcpConnections[topic] = make(map[string]*net.TCPConn)
		}
		tcpConnections[topic][addrRemote] = tcpConn
		peersConnected[topicip] = string(topic[:])
	}
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
		for topicip := range oldPeers {
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
