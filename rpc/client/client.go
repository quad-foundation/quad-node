package clientrpc

import (
	"github.com/quad-foundation/quad-node/tcpip"
	"log"
	"net/rpc"
	"strconv"
	"time"
)

const (
	retryInterval = 5 * time.Second
	bufferSize    = 1024 * 1024
)

var InRPC = make(chan []byte)
var OutRPC = make(chan []byte)

func ConnectRPC(ip string) {
	address := ip + ":" + strconv.Itoa(tcpip.Ports[tcpip.RPCTopic])
	var client *rpc.Client
	var err error
	for {
		client, err = rpc.Dial("tcp", address)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RPC server at %s: %v. Retrying in %v...", address, err, retryInterval)
		time.Sleep(retryInterval)
	}
	reply := make([]byte, bufferSize)
	for {
		select {
		case line := <-InRPC:
			err = client.Call("Listener.Send", line, &reply)
			if err != nil {
				log.Printf("RPC call failed: %v. Reconnecting...", err)
				for {
					client, err = rpc.Dial("tcp", address)
					if err == nil {
						break
					}
					log.Printf("Failed to reconnect to RPC server at %s: %v. Retrying in %v...", address, err, retryInterval)
					time.Sleep(retryInterval)
				}
			} else {
				OutRPC <- reply
			}
		}
	}
}
