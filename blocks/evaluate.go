package blocks

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	vm "github.com/quad-foundation/quad-node/core/evm"
	"github.com/quad-foundation/quad-node/core/stateDB"
	"github.com/quad-foundation/quad-node/params"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"log"
	"math"
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

func GenerateOptDataDEX(tx transactionsDefinition.Transaction, operation int) ([]byte, common.Address, int64, int64, float64, error) {
	// 2 - adding liquidity, 3 - buy trade, 4 -sell trade, 5 - withdraw token, 6 - withdraw QAD (5,6 inactive, just withdraw is selling opposite)
	amountToken := common.GetInt64FromByte(tx.TxData.OptData)
	sender := tx.TxParam.Sender
	tokenAddress := tx.ContractAddress
	if operation == 2 && (tx.TxData.Amount < 0 || amountToken < 0) || (operation == 3 || operation == 4) && (amountToken == 0 || tx.TxData.Amount != 0) || operation == 5 && amountToken == 0 || operation == 6 && tx.TxData.Amount == 0 {
		return nil, common.Address{}, 0, 0, 0, fmt.Errorf("withdraw one can perform on one currency the second should be 0")
	}

	accDex := account.GetDexAccountByAddressBytes(tokenAddress.GetBytes())
	price := float64(0)
	var amountCoinInt64, amountTokenInt64 int64
	balanceToken, err := GetBalance(tx.ContractAddress, sender)
	if err != nil {
		return nil, common.Address{}, 0, 0, 0, err
	}
	ba := [common.AddressLength]byte{}
	copy(ba[:], tx.ContractAddress.GetBytes())
	StateMutex.RLock()
	ti, ok := State.Tokens[ba]
	StateMutex.RUnlock()
	if !ok {
		return nil, common.Address{}, 0, 0, 0, fmt.Errorf("no token with a given address")
	}

	tokenPoolAmount := account.Int64toFloat64ByDecimals(accDex.TokenPool, ti.Decimals)
	coinPoolAmount := account.Int64toFloat64(accDex.CoinPool)
	amountTokenFloat := account.Int64toFloat64ByDecimals(amountToken, ti.Decimals)
	amountCoinFloat := account.Int64toFloat64ByDecimals(tx.TxData.Amount, common.Decimals)

	if coinPoolAmount > 0 && tokenPoolAmount > 0 {
		price = common.RoundCoin((tokenPoolAmount + amountTokenFloat) / (coinPoolAmount + amountCoinFloat))
	}

	switch operation {
	case 2: // add liquidity
		amountCoinInt64 = int64(-amountCoinFloat * math.Pow10(int(common.Decimals)))
		amountTokenInt64 = int64(-amountTokenFloat * math.Pow10(int(ti.Decimals)))
	case 5: // withdraw token
		if price > 0 {
			amount := common.RoundCoin(1.0 / price * amountTokenFloat)
			amountCoinInt64 = int64(amount * math.Pow10(int(common.Decimals)))
		} else {
			amountCoinInt64 = int64(0)
		}
		amountTokenInt64 *= -1
	case 6: // withdraw Coin
		if price > 0 {
			amount := common.RoundToken(price*amountCoinFloat, int(ti.Decimals))
			amountTokenInt64 = int64(amount * math.Pow10(int(ti.Decimals)))
		} else {
			amountTokenInt64 = int64(0)
		}
		amountCoinInt64 *= -1
	case 3: // buy
		if price > 0 {
			amount := common.RoundToken(price*amountCoinFloat, int(ti.Decimals))
			amountTokenInt64 = int64(amount * math.Pow10(int(ti.Decimals)))
		} else {
			amountTokenInt64 = 0
		}
		amountCoinInt64 *= -1
	case 4: // sell
		if price > 0 {
			amount := common.RoundCoin(1.0 / price * amountTokenFloat)
			amountCoinInt64 = int64(amount * math.Pow10(int(common.Decimals)))
		} else {
			amountCoinInt64 = 0
		}
		amountTokenInt64 *= -1
	default:
		return nil, common.Address{}, 0, 0, 0, fmt.Errorf("wrong operation on dex")
	}

	senderAccount := account.GetAccountByAddressBytes(tx.TxParam.Sender.GetBytes())
	if bytes.Compare(senderAccount.Address[:], tx.TxParam.Sender.GetBytes()) != 0 {
		return nil, common.Address{}, 0, 0, 0, fmt.Errorf("no account found in dex transfer")
	}

	if senderAccount.Balance+amountCoinInt64 < 0 {
		return nil, common.Address{}, 0, 0, 0, fmt.Errorf("not enough coins in account")
	}
	if balanceToken+amountTokenInt64 < 0 {
		return nil, common.Address{}, 0, 0, 0, fmt.Errorf("not enough tokens in account")
	}

	if accDex.Balances[senderAccount.Address].CoinBalance-amountCoinInt64 < 0 {
		return nil, common.Address{}, 0, 0, 0, fmt.Errorf("not enough coins in dex account")
	}
	if accDex.Balances[senderAccount.Address].TokenBalance-amountTokenInt64 < 0 {
		return nil, common.Address{}, 0, 0, 0, fmt.Errorf("not enough tokens in dex account")
	}

	var fromAccountAddress common.Address
	var optData []byte

	if amountTokenInt64 > 0 {
		dexByte := common.LeftPadBytes(senderAccount.Address[:], 32)
		amountByte := common.LeftPadBytes(common.Int64ToBytes(amountTokenInt64), 32)
		optData = append(stateDB.TransferFunc, dexByte...)
		optData = append(optData, amountByte...)
		fromAccountAddress = tx.ContractAddress
	} else if amountTokenInt64 < 0 {
		dexByte := common.LeftPadBytes(tx.ContractAddress.GetBytes(), 32)
		amountByte := common.LeftPadBytes(common.Int64ToBytes(-amountTokenInt64), 32)
		optData = append(stateDB.TransferFunc, dexByte...)
		optData = append(optData, amountByte...)
		fromAccountAddress = sender
	}
	log.Println(common.Bytes2Hex(optData))
	return optData, fromAccountAddress, amountCoinInt64, amountTokenInt64, price, nil
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
		n, err := account.IntDelegatedAccountFromAddress(addressRecipient)
		if err == nil && n > 512 { // 514 == operation 2 etc...
			//DEX checking transaction
			dexOptData, fromAddress, coinAmount, tokenAmount, price, err := GenerateOptDataDEX(t, n-512)
			log.Printf("Token Price: %v\n", price)
			if err != nil {
				log.Println(err)
				return false, nil, nil, nil, nil
			}
			// transfering tokens
			l, _, _, _, err := EvaluateSCDex(t.ContractAddress, fromAddress, dexOptData, t, bl)
			if err != nil {
				log.Println(err)
				return false, logs, map[[common.HashLength]byte]common.Address{}, map[[common.AddressLength]byte][]byte{}, map[[common.HashLength]byte][]byte{}
			}
			t.OutputLogs = []byte(l)
			err = t.StoreToDBPoolTx(poolprefix)
			if err != nil {
				log.Println(err)
				return false, logs, map[[common.HashLength]byte]common.Address{}, map[[common.AddressLength]byte][]byte{}, map[[common.HashLength]byte][]byte{}
			}
			aa := [common.AddressLength]byte{}
			copy(aa[:], t.TxParam.Sender.GetBytes())
			// transfering coins QAD
			if coinAmount < 0 {
				err = AddBalance(aa, coinAmount)
				if err != nil {
					return false, nil, nil, nil, nil
				}
			} else {

			}

			ba := [common.AddressLength]byte{}
			copy(ba[:], t.ContractAddress.GetBytes())
			StateMutex.RLock()
			ti, ok := State.Tokens[ba]
			StateMutex.RUnlock()
			if !ok {
				log.Println("no token with a given address")
				return false, nil, nil, nil, nil
			}

			accDex := account.GetDexAccountByAddressBytes(t.ContractAddress.GetBytes())

			accDex.TokenPrice = int64(price * math.Pow10(int(common.Decimals+ti.Decimals)))
			accDex.TokenPool += tokenAmount
			accDex.CoinPool += coinAmount
			coinAmountTmp := accDex.Balances[aa].CoinBalance + coinAmount
			tokenAmountTmp := accDex.Balances[aa].TokenBalance + tokenAmount
			balances := accDex.Balances
			if balances == nil {
				balances = make(map[[common.AddressLength]byte]account.CoinTokenDetails)
			}
			balances[aa] = account.CoinTokenDetails{
				CoinBalance:  coinAmountTmp,
				TokenBalance: tokenAmountTmp,
			}
			accDex.Balances = balances
			account.SetDexAccountByAddressBytes(t.ContractAddress.GetBytes(), accDex)

			continue
		}
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
		ret, leftOverGas, err = VM.Call(vm.AccountRef(origin), address, code, uint64(tx.GasUsage)*uint64(gasMult), new(big.Int).SetInt64(0))
		if err != nil {
			return logger.ToString(), ret, address, leftOverGas, err
		}
	}

	return logger.ToString(), ret, address, uint64(float64(leftOverGas) / gasMult), nil
}

func EvaluateSCDex(tokenAddress common.Address, sender common.Address, optData []byte, tx transactionsDefinition.Transaction, bl Block) (logs string, ret []byte, address common.Address, leftOverGas uint64, err error) {

	gasMult := 10.0

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

	VM.Origin = sender
	VM.GasPrice = new(big.Int).SetInt64(0)

	ret, leftOverGas, err = VM.Call(vm.AccountRef(sender), tokenAddress, optData, uint64(tx.GasUsage)*uint64(gasMult), new(big.Int).SetInt64(0))
	if err != nil {
		return logger.ToString(), ret, tokenAddress, leftOverGas, err
	}

	return logger.ToString(), ret, tokenAddress, uint64(float64(leftOverGas) / gasMult), nil
}

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
