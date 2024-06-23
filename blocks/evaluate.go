package blocks

import (
	"bytes"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	vm "github.com/quad-foundation/quad-node/core/evm"
	"github.com/quad-foundation/quad-node/core/stateDB"
	"github.com/quad-foundation/quad-node/params"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"log"
	"math/big"
	"sync"
)

var State stateDB.StateAccount
var StateMutex sync.RWMutex
var VM *vm.EVM

type PasiveFunction struct {
	Address common.Address `json:"address"`
	OptData []byte         `json:"optData"`
	Height  int64          `json:"height"`
}

func init() {
	State = stateDB.CreateStateDB()
}

func EvaluateSCForBlock(bl Block) (bool, map[[common.HashLength]byte]string, map[[common.HashLength]byte]common.Address, map[[common.AddressLength]byte][]byte, map[[common.HashLength]byte][]byte) {
	addresses := map[[common.HashLength]byte]common.Address{}
	logs := map[[common.HashLength]byte]string{}
	rets := map[[common.HashLength]byte][]byte{}
	optDatas := map[[common.AddressLength]byte][]byte{}
	for _, th := range bl.GetBlockTransactionsHashes() {
		poolprefix := common.TransactionPoolHashesDBPrefix[:]
		t, err := transactionsDefinition.LoadFromDBPoolTx(poolprefix, th.GetBytes())
		if err != nil {
			poolprefix = common.TransactionDBPrefix[:]
			t, err = transactionsDefinition.LoadFromDBPoolTx(poolprefix, th.GetBytes())
			if err != nil {
				log.Println(err)
				return false, logs, map[[common.HashLength]byte]common.Address{}, map[[common.AddressLength]byte][]byte{}, map[[common.HashLength]byte][]byte{}
			}
		}
		addressRecipient := t.TxData.Recipient
		_, err = account.IntDelegatedAccountFromAddress(addressRecipient)
		if err == nil {
			continue
		}
		if len(t.TxData.OptData) == 0 {
			continue
		}
		l, ret, address, _, err := EvaluateSC(t, bl)
		if t.TxData.Recipient == common.EmptyAddress() {
			code := t.TxData.OptData
			if ok := IsTokenToRegister(code); ok && err == nil {
				input := stateDB.NameFunc
				output, _, _, _, _, err := GetViewFunctionReturns(address, input, bl)
				var name string
				if err == nil {
					name = common.GetStringFromSCBytes(common.Hex2Bytes(output), 0)
				}
				input = stateDB.SymbolFunc
				output, _, _, _, _, err = GetViewFunctionReturns(address, input, bl)
				var symbol string
				if err == nil {
					symbol = common.GetStringFromSCBytes(common.Hex2Bytes(output), 0)
				}
				input = stateDB.DecimalsFunc
				output, _, _, _, _, err = GetViewFunctionReturns(address, input, bl)
				var decimals uint8
				if err == nil {
					decimals = uint8(common.GetUintFromSCByte(common.Hex2Bytes(output)))
				}
				StateMutex.Lock()
				State.RegisterNewToken(address, name, symbol, decimals)
				StateMutex.Unlock()
			}
		}
		if err != nil {
			return false, logs, map[[common.HashLength]byte]common.Address{}, map[[common.AddressLength]byte][]byte{}, map[[common.HashLength]byte][]byte{}
		}
		//TODO we should refund left gas
		//t.GasUsage -= int64(leftOverGas)
		t.ContractAddress = address
		outputLogs := []byte(l)
		//if err != nil {
		//	log.Println(err)
		//	return false, logs, map[[common.HashLength]byte]common.Address{}, map[[common.AddressLength]byte][]byte{}, map[[common.HashLength]byte][]byte{}
		//}
		t.OutputLogs = outputLogs[:]
		err = t.StoreToDBPoolTx(poolprefix)
		if err != nil {
			log.Println(err)
			return false, logs, map[[common.HashLength]byte]common.Address{}, map[[common.AddressLength]byte][]byte{}, map[[common.HashLength]byte][]byte{}
		}
		hh := [common.HashLength]byte{}
		copy(hh[:], t.Hash.GetBytes()[:])
		rets[hh] = ret
		addresses[hh] = address
		logs[hh] = l
		aa := [common.AddressLength]byte{}
		copy(aa[:], address.GetBytes()[:])
		optDatas[aa] = t.TxData.OptData
	}
	return true, logs, addresses, optDatas, rets
}

func EvaluateSC(tx transactionsDefinition.Transaction, bl Block) (logs string, ret []byte, address common.Address, leftOverGas uint64, err error) {
	if len(tx.TxData.OptData) == 0 {
		log.Println("no smart contract in transaction")
		return logs, ret, address, leftOverGas, nil
	}
	gasMult := 10.0

	origin := tx.TxParam.Sender
	code := tx.TxData.OptData
	blockCtx := vm.BlockContext{
		CanTransfer: nil,
		Transfer:    nil,
		GetHash:     func(height uint64) common.Hash { return bl.GetBlockHash() },
		Coinbase:    common.EmptyAddress(),
		GasLimit:    uint64(common.MaxGasUsage) * uint64(gasMult),
		BlockNumber: new(big.Int).SetInt64(bl.GetHeader().Height),
		Time:        new(big.Int).SetInt64(common.GetCurrentTimeStampInSecond()),
		Difficulty:  new(big.Int).SetInt64(int64(bl.GetHeader().Difficulty)),
		BaseFee:     new(big.Int).SetInt64(int64(0)),
		Random:      nil,
	}
	logger := vm.CreateGVMLogger()
	jumpTable := vm.GetGenericJumpTable()

	configCtx := vm.Config{
		Debug:                   true,
		Tracer:                  &logger,
		NoBaseFee:               true,
		EnablePreimageRecording: true,
		JumpTable:               &jumpTable,
		ExtraEips:               []int{},
	}
	txCtx := vm.TxContext{
		Origin:   tx.TxParam.Sender,
		GasPrice: new(big.Int).SetInt64(tx.GasPrice),
	}
	StateMutex.Lock()
	defer StateMutex.Unlock()

	VM = vm.NewEVM(blockCtx, txCtx, &State, params.AllEthashProtocolChanges, configCtx)
	defer VM.Cancel()

	VM.Origin = origin
	VM.GasPrice = new(big.Int).SetInt64(tx.GasPrice)
	nonce := new(big.Int).SetInt64(int64(tx.TxParam.Nonce))

	if tx.TxData.Recipient == common.EmptyAddress() {
		ret, address, leftOverGas, err = VM.Create(vm.AccountRef(origin), code, uint64(tx.GasUsage)*uint64(gasMult), nonce)

		if err != nil {
			return logger.ToString(), ret, address, leftOverGas, err
		}
	} else {
		address = tx.TxData.Recipient
		ret, leftOverGas, err = VM.Call(vm.AccountRef(origin), address, code, uint64(tx.GasUsage), new(big.Int).SetInt64(0))
		if err != nil {
			return logger.ToString(), ret, address, leftOverGas, err
		}
	}

	return logger.ToString(), ret, address, uint64(float64(leftOverGas) / gasMult), nil
}

//func EvaluateSCDex(tokenAddress common.Address, sender common.Address, optData []byte, tx stakingDefinition.StakeTransaction, bl block.BlockMainChain) (logs string, ret []byte, address common.Address, leftOverGas uint64, err error) {
//
//	gasMult := 10.0
//
//	blockCtx := vm.BlockContext{
//		CanTransfer: nil,
//		Transfer:    nil,
//		GetHash:     func(height uint64) common.Hash { return bl.GetGVMHash() },
//		Coinbase:    common.EmptyAddress(),
//		GasLimit:    uint64(common.MaxGasUsage) * uint64(gasMult),
//		BlockNumber: new(big.Int).SetInt64(bl.GetHeader().Height),
//		Time:        new(big.Int).SetInt64(common.GetCurrentTimeStampInSecond()),
//		Difficulty:  new(big.Int).SetInt64(int64(bl.GetHeader().Difficulty)),
//		BaseFee:     new(big.Int).SetInt64(int64(0)),
//		Random:      nil,
//	}
//	logger := vm.CreateGVMLogger()
//	jumpTable := vm.GetGenericJumpTable()
//
//	configCtx := vm.Config{
//		Debug:                   true,
//		Tracer:                  &logger,
//		NoBaseFee:               true,
//		EnablePreimageRecording: true,
//		JumpTable:               &jumpTable,
//		ExtraEips:               []int{},
//	}
//	txCtx := vm.TxContext{
//		Origin:   tx.Sender,
//		GasPrice: new(big.Int).SetInt64(0),
//	}
//	StateMutex.Lock()
//	defer StateMutex.Unlock()
//
//	VM = vm.NewEVM(blockCtx, txCtx, &State, params.AllEthashProtocolChanges, configCtx)
//	defer VM.Cancel()
//
//	VM.Origin = sender
//	VM.GasPrice = new(big.Int).SetInt64(0)
//
//	ret, leftOverGas, err = VM.Call(vm.AccountRef(sender), tokenAddress, optData, uint64(210000), new(big.Int).SetInt64(0))
//	if err != nil {
//		return logger.ToString(), ret, tokenAddress, leftOverGas, err
//	}
//
//	return logger.ToString(), ret, tokenAddress, uint64(float64(leftOverGas) / gasMult), nil
//}

func GetViewFunctionReturns(contractAddr common.Address, OptData []byte, bl Block) (outputs string, logs string, ret []byte, address common.Address, leftOverGas uint64, err error) {

	origin := common.EmptyAddress()
	input := OptData
	blockCtx := vm.BlockContext{
		CanTransfer: nil,
		Transfer:    nil,
		GetHash:     func(height uint64) common.Hash { return bl.GetBlockHash() },
		Coinbase:    common.EmptyAddress(),
		GasLimit:    uint64(common.MaxGasUsage),
		BlockNumber: new(big.Int).SetInt64(bl.GetHeader().Height),
		Time:        new(big.Int).SetInt64(common.GetCurrentTimeStampInSecond()),
		Difficulty:  new(big.Int).SetInt64(int64(bl.GetHeader().Difficulty)),
		BaseFee:     new(big.Int).SetInt64(int64(0)),
		Random:      nil,
	}
	logger := vm.CreateGVMLogger()
	jumpTable := vm.GetGenericJumpTable()

	configCtx := vm.Config{
		Debug:                   true,
		Tracer:                  &logger,
		NoBaseFee:               true,
		EnablePreimageRecording: true,
		JumpTable:               &jumpTable,
		ExtraEips:               []int{},
	}
	txCtx := vm.TxContext{
		Origin:   origin,
		GasPrice: new(big.Int).SetInt64(0),
	}
	StateMutex.Lock()
	defer StateMutex.Unlock()
	VM = vm.NewEVM(blockCtx, txCtx, &State, params.AllEthashProtocolChanges, configCtx)
	defer VM.Cancel()

	VM.Origin = origin
	VM.GasPrice = new(big.Int).SetInt64(common.MaxGasUsage)
	ret, leftOverGas, err = VM.StaticCall(vm.AccountRef(origin), contractAddr, input, uint64(common.MaxGasUsage))
	if err != nil {
		return logger.Output, logs, ret, address, leftOverGas, err
	}

	return logger.Output, logger.ToString(), ret, address, leftOverGas, nil
}

func IsTokenToRegister(code []byte) bool {
	toRegister := true
	if bytes.Index(code, stateDB.NameFunc) < 0 {
		toRegister = false
	}
	if bytes.Index(code, stateDB.BalanceOfFunc) < 0 {
		toRegister = false
	}
	if bytes.Index(code, stateDB.TransferFunc) < 0 {
		toRegister = false
	}
	if bytes.Index(code, stateDB.SymbolFunc) < 0 {
		toRegister = false
	}
	if bytes.Index(code, stateDB.DecimalsFunc) < 0 {
		toRegister = false
	}
	return toRegister
}

func GetBalance(coin common.Address, owner common.Address) (int64, error) {

	inputs := stateDB.BalanceOfFunc
	ba := common.LeftPadBytes(owner.GetBytes(), 32)
	inputs = append(inputs, ba...)

	h := common.GetHeight()

	var bl Block
	var err error

	bl, err = LoadBlock(h - 1)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	output, _, _, _, _, err := GetViewFunctionReturns(coin, inputs, bl)
	if err != nil {
		log.Println("Some error in SC query Get Balance", err)
		return 0, err
	}
	if output != "" {
		bal := common.GetInt64FromSCByte(common.Hex2Bytes(output))
		return bal, nil
	} else {
		return 0, nil
	}
}
