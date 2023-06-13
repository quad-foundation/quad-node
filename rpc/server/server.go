package serverrpc

import (
	"encoding/json"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/services/transactionServices"
	"github.com/chainpqc/chainpqc-node/statistics"
	"github.com/chainpqc/chainpqc-node/tcpip"
	"github.com/chainpqc/chainpqc-node/wallet"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"sync"
)

var listenerMutex sync.Mutex

type Listener []byte

func ListenRPC() {
	var address = "0.0.0.0:" + strconv.Itoa(tcpip.Ports[tcpip.RPCTopic])
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error resolving TCP address: %v", err)
	}
	defer listener.Close()
	err = rpc.Register(new(Listener))
	if err != nil {
		log.Fatalf("Error registering RPC listener: %v", err)
	}
	log.Printf("RPC server listening on %s", address)
	rpc.Accept(listener)
}

func (l *Listener) Send(line []byte, reply *[]byte) error {
	listenerMutex.Lock()
	defer listenerMutex.Unlock()
	if len(line) < 4 {
		*reply = []byte("Error with message. Too small length calling server")
		return nil
	}
	operation := string(line[:4])
	byt := line[4:]
	switch operation {
	case "STAT":
		handleSTAT(byt, reply)
	case "WALL":
		handleWALL(byt, reply)
	case "TRAN":
		handleTRAN(byt, reply)
	//case "VIEW":
	//	handleVIEW(byt, reply)
	//case "ACCT":
	//	handleACCT(byt, reply)
	//case "DETS":
	//	handleDETS(byt, reply)
	//case "STAK":
	//	handleSTAK(byt, reply)
	//case "ACCS":
	//	handleACCS(reply)
	//case "LTKN":
	//	handleLTKN(reply)
	//case "GTBL":
	//	handleGTBL(byt, reply)
	default:
		*reply = []byte("Invalid operation")
	}
	return nil
}

//	func handleSTAT(reply *[]byte) {
//		st := statistics.MainStats{}
//		st, err := st.LoadStats()
//		if err != nil {
//			log.Println("Cannot update stats")
//			return
//		}
//		r, err := json.Marshal(st)
//		if err != nil {
//			log.Println("Cannot marshal stat's struct")
//			return
//		}
//		*reply = r
//	}
func handleWALL(line []byte, reply *[]byte) {
	log.Println(string(line))
	w := wallet.GetActiveWallet()
	r, err := json.Marshal(w)
	if err != nil {
		log.Println("Cannot marshal stat's struct")
		return
	}
	*reply = r
}
func handleTRAN(byt []byte, reply *[]byte) {

	*reply = []byte("transaction sent")
	transactionServices.OnMessage("toSend", byt)

}
func handleSTAT(byt []byte, reply *[]byte) {
	st, err := statistics.LoadStats()
	if err != nil {
		log.Println("Can't update stats")
		return
	}
	msb, err := common.Marshal(st.MainStats, common.StatDBPrefix)
	if err != nil {
		log.Println(err)
		return
	}
	*reply = msb
}
