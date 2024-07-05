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
	PeersCount       = 0
	receiveMutex     = &sync.Mutex{}
	waitChan         = make(chan []byte)
	tcpConnections   = make(map[[2]byte]map[string]*net.TCPConn)
	PeersMutex       = &sync.RWMutex{}
	Quit             chan os.Signal
	TransactionTopic = [2]byte{'T', 'T'}
	NonceTopic       = [2]byte{'N', 'N'}
	SelfNonceTopic   = [2]byte{'S', 'S'}
	SyncTopic        = [2]byte{'B', 'B'}
	RPCTopic         = [2]byte{'R', 'P'}
)
var Ports = map[[2]byte]int{
	TransactionTopic: 9091,
	NonceTopic:       8091,
	SelfNonceTopic:   7091,
	SyncTopic:        6091,
	RPCTopic:         9009,
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

	if conn == nil {
		return []byte("<-CLS->")
	}

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

	PeersMutex.Lock()
	defer PeersMutex.Unlock()

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

func GetPeersConnected() map[string]string {
	PeersMutex.RLock()
	defer PeersMutex.RUnlock()

	copyOfPeers := make(map[string]string, len(peersConnected))
	for key, value := range peersConnected {
		copyOfPeers[key] = value
	}

	return copyOfPeers
}

func GetIPsConnected() [][]byte {
	PeersMutex.RLock()
	defer PeersMutex.RUnlock()
	uniqueIPs := make(map[string]struct{})
	for key, value := range peersConnected {
		if value == "NN" {
			ip := key[2:]
			uniqueIPs[ip] = struct{}{}
		}
	}
	var ips [][]byte
	for ip := range uniqueIPs {
		parsedIP := net.ParseIP(ip).To4()
		if parsedIP != nil {
			ips = append(ips, parsedIP)
		}
	}
	return ips
}

func GetIPsfrombytes(peers [][]byte) []string {
	var ips []string
	for _, ipb := range peers {
		if len(ipb) == 4 {
			ip := fmt.Sprintf("%d.%d.%d.%d", ipb[0], ipb[1], ipb[2], ipb[3])
			ips = append(ips, ip)
		}
	}
	return ips
}

func AddNewPeer(peer string, topic [2]byte) {
	PeersMutex.Lock()
	defer PeersMutex.Unlock()
	topicip := string(topic[:]) + peer
	peersConnected[topicip] = string(topic[:])
}

func GetPeersCount() int {
	PeersMutex.RLock()
	defer PeersMutex.RUnlock()
	return PeersCount
}

func LookUpForNewPeersToConnect(chanPeer chan string) {
	for {
		PeersMutex.RLock()
		for topicip, topic := range peersConnected {
			_, ok := oldPeers[topicip]
			if ok == false {
				log.Println("Found new peer with ip", topicip)
				oldPeers[topicip] = topic
				chanPeer <- topicip
				PeersCount = len(GetIPsConnected())
			}
		}
		for topicip := range oldPeers {
			_, ok := peersConnected[topicip]
			if ok == false {
				log.Println("New peer is deleted with ip", topicip)
				delete(oldPeers, topicip)
			}
		}
		PeersMutex.RUnlock()
		time.Sleep(time.Second)
	}
}
