package serverrpc

import (
	"encoding/json"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/blocks"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/services/transactionServices"
	"github.com/quad-foundation/quad-node/statistics"
	"github.com/quad-foundation/quad-node/tcpip"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/quad-foundation/quad-node/wallet"
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
	case "ACCT":
		handleACCT(byt, reply)
	//case "ACCS":
	//	handleACCS(byt, reply)
	case "DETS":
		handleDETS(byt, reply)
	case "STAK":
		handleSTAK(byt, reply)
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

func handleDETS(line []byte, reply *[]byte) {

	switch len(line) {
	case common.AddressLength:
		byt := [common.AddressLength]byte{}
		copy(byt[:], line)
		account.AccountsRWMutex.RLock()
		acc := account.Accounts.AllAccounts[byt]
		account.AccountsRWMutex.RUnlock()
		am := acc.Marshal()
		*reply = append([]byte("AC"), am...)
		break
	case common.HashLength:
		tx, err := transactionsDefinition.LoadFromDBPoolTx(common.TransactionDBPrefix[:], line)
		if err != nil {
			log.Println(err)
			*reply = []byte("TX")
			return
		}
		txb := tx.GetBytes()
		*reply = append([]byte("TX"), txb...)
		break
	case 8:
		height := common.GetInt64FromByte(line)
		block, err := blocks.LoadBlock(height)
		if err != nil {
			log.Println(err)
			*reply = []byte("BL")
			return
		}
		bb := block.GetBytes()
		*reply = append([]byte("BL"), bb...)
		break
	default:
		*reply = []byte("NO")
	}
}

func handleACCT(line []byte, reply *[]byte) {

	byt := [common.AddressLength]byte{}
	copy(byt[:], line[:common.AddressLength])
	account.AccountsRWMutex.RLock()
	acc := account.Accounts.AllAccounts[byt]
	account.AccountsRWMutex.RUnlock()
	am := acc.Marshal()
	*reply = am
}

func handleSTAK(line []byte, reply *[]byte) {

	byt := [common.AddressLength]byte{}
	copy(byt[:], line[:common.AddressLength])
	n := int(line[common.AddressLength])
	account.StakingRWMutex.RLock()
	acc := account.StakingAccounts[n].AllStakingAccounts[byt]
	account.StakingRWMutex.RUnlock()
	am := acc.Marshal()
	*reply = am
}

//func handleACCS(line []byte, reply *[]byte) {
//
//	byt := [common.AddressLength]byte{}
//	copy(byt[:], line[:common.AddressLength])
//	for i:=0;i<256;i++ {
//		if common.ContainsKeyInMap(account.StakingAccounts[i].AllStakingAccounts, byt) {
//			acc := account.StakingAccounts[i].AllStakingAccounts[byt]
//			am := acc.Marshal()
//		}
//	}
//	*reply = am
//}

func handleTRAN(byt []byte, reply *[]byte) {

	*reply = []byte("transaction sent")
	transactionServices.OnMessage("toSend", byt)

}
func handleSTAT(byt []byte, reply *[]byte) {
	statistics.GmsMutex.Mutex.Lock()
	defer statistics.GmsMutex.Mutex.Unlock()
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
