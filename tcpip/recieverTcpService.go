package tcpip

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	peersConnected   = map[[6]byte][2]byte{}
	oldPeers         = map[[6]byte][2]byte{}
	PeersCount       = 0
	receiveMutex     = &sync.Mutex{}
	waitChan         = make(chan []byte)
	tcpConnections   = make(map[[2]byte]map[[4]byte]*net.TCPConn)
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

var MyIP [4]byte

func init() {
	Quit = make(chan os.Signal)
	signal.Notify(Quit, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	MyIP = getInternalIp()
	for k := range Ports {
		tcpConnections[k] = map[[4]byte]*net.TCPConn{}
	}
}

func getInternalIp() [4]byte {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println("Can not obtain net interface")
		return [4]byte{}
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Println("Can not get net addresses")
			return [4]byte{}
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
				return [4]byte(ip.To4())
			}
		}
	}
	return [4]byte{}
}
func Listen(ip [4]byte, port int) (*net.TCPListener, error) {
	ipport := fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], port)
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

func Send(conn *net.TCPConn, message []byte) error {
	message = append(message, []byte("<-END->")...)
	_, err := conn.Write(message)
	if err != nil {
		log.Printf("Can't send response: %v", err)
		return err
	}
	return nil
}

// Receive reads data from the connection and handles errors
func Receive(topic [2]byte, conn *net.TCPConn) []byte {
	const bufSize = 1024 //1048576

	if conn == nil {
		return []byte("<-CLS->")
	}

	buf := make([]byte, bufSize)
	n, err := conn.Read(buf)

	if err != nil {
		//handleConnectionError(err, topic, conn)
		conn.Close()
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
	ips := strings.Split(ra[0], ".")
	var ip [4]byte
	for i := 0; i < 4; i++ {
		num, err := strconv.Atoi(ips[i])
		if err != nil {
			fmt.Println("Invalid IP address segment:", ips[i])
			return
		}
		ip[i] = byte(num)
	}
	var topicipBytes [6]byte
	var addrRemoteBytes [4]byte
	copy(topicipBytes[:], append(topic[:], ip[:]...))
	copy(addrRemoteBytes[:], ip[:])
	PeersMutex.RLock()
	if _, ok := peersConnected[topicipBytes]; !ok {
		log.Println("New connection from address", ra[0], "on topic", topic)
		PeersMutex.RUnlock()
		PeersMutex.Lock()
		// Initialize the map for the topic if it doesn't exist
		if _, ok := tcpConnections[topic]; !ok {
			tcpConnections[topic] = make(map[[4]byte]*net.TCPConn)
		}
		tcpConnections[topic][addrRemoteBytes] = tcpConn
		peersConnected[topicipBytes] = topic
		PeersMutex.Unlock()
	} else {
		PeersMutex.RUnlock()
	}
}

func GetPeersConnected(topic [2]byte) map[[6]byte][2]byte {
	PeersMutex.RLock()
	defer PeersMutex.RUnlock()

	copyOfPeers := make(map[[6]byte][2]byte, len(peersConnected))
	for key, value := range peersConnected {
		if value == topic {
			copyOfPeers[key] = value
		}
	}

	return copyOfPeers
}

func GetIPsConnected() [][]byte {
	uniqueIPs := make(map[[4]byte]struct{})
	ipb := [4]byte{}
	for key, value := range peersConnected {
		if value == [2]byte{'N', 'N'} {
			copy(ipb[:], key[2:])
			if bytes.Equal(ipb[:], MyIP[:]) {
				continue
			}
			uniqueIPs[ipb] = struct{}{}
		}
	}
	var ips [][]byte
	for ip := range uniqueIPs {
		ips = append(ips, ip[:])
	}
	return ips
}

func GetPeersCount() int {
	PeersMutex.RLock()
	defer PeersMutex.RUnlock()
	return PeersCount
}

func LookUpForNewPeersToConnect(chanPeer chan []byte) {
	for {
		PeersMutex.Lock()
		for topicip, topic := range peersConnected {
			_, ok := oldPeers[topicip]
			if ok == false {
				log.Println("Found new peer with ip", topicip)
				oldPeers[topicip] = topic
				chanPeer <- topicip[:]
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
		PeersMutex.Unlock()
		time.Sleep(time.Second * 10)
	}
}
