package vm

import (
	"errors"
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
	ethc "github.com/ethereum/go-ethereum/core"
	ethvm "github.com/ethereum/go-ethereum/core/vm"

	"github.com/sudachen/playground/libeth/state"
	"github.com/sudachen/playground/libeth"

	"github.com/sudachen/playground/crypto"
	"github.com/ethereum/go-ethereum/common"
)

type message struct {
	from              eth.Address
	to                *eth.Address
	value, gas, price *big.Int
	data              []byte
	nonce             uint64
}

func NewMessage(from eth.Address, to *eth.Address, data []byte, value, gas, price *big.Int, nonce uint64) message {
	return message{from, to, value, gas, price, data, nonce}
}

func (m message) Hash() []byte                       { return nil }
func (m message) From() (eth.Address, error)         { return m.from, nil }
func (m message) FromFrontier() (eth.Address, error) { return m.from, nil }
func (m message) To() *eth.Address                   { return m.to }
func (m message) GasPrice() *big.Int                 { return m.price }
func (m message) Gas() *big.Int                      { return m.gas }
func (m message) Value() *big.Int                    { return m.value }
func (m message) Nonce() uint64                      { return m.nonce }
func (m message) Data() []byte                       { return m.data }

func comHash(a eth.Hash) (e common.Hash) {
	copy(e[:], a[:])
	return
}

func ethHash(a common.Hash) (e eth.Hash) {
	copy(e[:], a[:])
	return
}

func comAddress(a eth.Address) (e common.Address) {
	copy(e[:], a[:])
	return
}

func ethAddress(a common.Address) (e eth.Address) {
	copy(e[:], a[:])
	return
}

func ethAddressOpt(a *common.Address) *eth.Address {
	if a == nil {
		return nil
	}
	e := ethAddress(*a)
	return &e
}

var defaultRuleset = &libeth.RuleSet{
	HomesteadBlock:           big.NewInt(1150000),
	HomesteadGasRepriceBlock: big.NewInt(2500000),
	DiehardBlock:             big.NewInt(3000000),
	ExplosionBlock:           big.NewInt(5000000),
}

type RuleSet struct {
	*libeth.RuleSet
}

func (r RuleSet) IsHomestead(n *big.Int) bool {
	return n.Cmp(r.RuleSet.HomesteadBlock) >= 0
}

func (r RuleSet) GasTable(num *big.Int) *ethvm.GasTable {
	if r.RuleSet.HomesteadGasRepriceBlock == nil || num == nil || num.Cmp(r.RuleSet.HomesteadGasRepriceBlock) < 0 {
		return &ethvm.GasTable{
			ExtcodeSize:     big.NewInt(20),
			ExtcodeCopy:     big.NewInt(20),
			Balance:         big.NewInt(20),
			SLoad:           big.NewInt(50),
			Calls:           big.NewInt(40),
			Suicide:         big.NewInt(0),
			ExpByte:         big.NewInt(10),
			CreateBySuicide: nil,
		}
	}
	if r.RuleSet.DiehardBlock == nil || num == nil || num.Cmp(r.RuleSet.DiehardBlock) < 0 {
		return &ethvm.GasTable{
			ExtcodeSize:     big.NewInt(700),
			ExtcodeCopy:     big.NewInt(700),
			Balance:         big.NewInt(400),
			SLoad:           big.NewInt(200),
			Calls:           big.NewInt(700),
			Suicide:         big.NewInt(5000),
			ExpByte:         big.NewInt(10),
			CreateBySuicide: big.NewInt(25000),
		}
	}

	return &ethvm.GasTable{
		ExtcodeSize:     big.NewInt(700),
		ExtcodeCopy:     big.NewInt(700),
		Balance:         big.NewInt(400),
		SLoad:           big.NewInt(200),
		Calls:           big.NewInt(700),
		Suicide:         big.NewInt(5000),
		ExpByte:         big.NewInt(50),
		CreateBySuicide: big.NewInt(25000),
	}
}

type nvm struct {
	db       *database
	origin   eth.Address
	coinbase eth.Address

	depth        int
	skipTransfer bool
	initial      bool
	Gas          *big.Int
	number       *big.Int
	time         *big.Int
	difficulty   *big.Int
	gasLimit     *big.Int

	//parent   common.Hash

	vmTest bool
	evm    *ethvm.EVM

	blockhash func(*big.Int) common.Hash
	rules     RuleSet
}

type database struct {
	libeth.MutableState
	Refund *big.Int
}

func NewVM() common.VM {
	return &nvm{}
}

func (vm *nvm) Execute(tx *libeth.Transaction, bi *libeth.BlockInfo, st libeth.State) (
	/*out*/ []byte,
	/*usedGas*/ *big.Int,
	/*resultState*/ libeth.State,
	/*executionError*/ error) {

	db := state.NewMicroState(st)
	snapshot := db.Snapshot()
	message := NewMessage(ethAddress(tx.From), ethAddressOpt(tx.To), tx.Data, tx.Value, tx.GasLimit, tx.GasPrice, tx.Nonce)

	vm.origin = ethAddress(tx.From)
	vm.coinbase = ethAddress(bi.Coinbase)
	vm.number = bi.Number
	vm.blockhash = bi.Blockhash
	vm.difficulty = bi.Difficulty
	vm.gasLimit = bi.GasLimit
	vm.time = bi.Time

	vm.rules = RuleSet{bi.RuleSet}
	if vm.rules.RuleSet == nil {
		vm.rules.RuleSet = defaultRuleset
	}

	vm.db = &database{db, new(big.Int)}
	vm.Gas = new(big.Int)
	vm.evm = ethvm.New(vm)

	gaspool := new(ethc.GasPool).AddGas(bi.GasLimit)

	out, usedGas, err := ethc.ApplyMessage(vm, message, gaspool)

	if ethc.IsNonceErr(err) || ethc.IsInvalidTxErr(err) || ethc.IsGasLimitErr(err) {
		db.Revert(snapshot)
	}

	return out, usedGas, db.Freeze(), err
}

type account struct {
	libeth.MutableAccount
}

func (a *account) SubBalance(amount *big.Int)         { a.MutableAccount.SubBalance(amount) }
func (a *account) AddBalance(amount *big.Int)         { a.MutableAccount.AddBalance(amount) }
func (a *account) SetBalance(amount *big.Int)         { a.MutableAccount.SetBalance(amount) }
func (a *account) SetNonce(nonce uint64)              { a.MutableAccount.SetNonce(nonce) }
func (a *account) Balance() *big.Int                  { return a.MutableAccount.Balance() }
func (a *account) SetCode(hash eth.Hash, code []byte) { a.MutableAccount.SetCode(code) }
func (a *account) Address() eth.Address               { return ethAddress(a.MutableAccount.Address) }
func (a *account) ReturnGas(*big.Int, *big.Int)       {}
func (a *account) Value() *big.Int                    { return nil }

var stopError = errors.New("stop")

func (a *account) ForEachStorage(cb func(key, value eth.Hash) bool) {
	a.MutableAccount.ProcessValues(
		func(key, value common.Hash) error {
			if cb(ethHash(key), ethHash(value)) {
				return nil
			} else {
				return stopError
			}
		},
		false,
	)
}

func (db *database) GetOrNew(a eth.Address) ethvm.Account {
	address := comAddress(a)
	return &account{common.MutableAccount{State: db.MutableState, Address: address}}
}

func (db *database) GetAccount(a eth.Address) ethvm.Account {
	address := comAddress(a)
	if db.MutableState.Exists(address) {
		return &account{common.MutableAccount{State: db.MutableState, Address: address}}
	}
	return nil
}

func (db *database) CreateAccount(a eth.Address) ethvm.Account {
	address := comAddress(a)
	db.MutableState.Create(address)
	return &account{common.MutableAccount{State: db.MutableState, Address: address}}
}

func (db *database) AddBalance(a eth.Address, v *big.Int) {
	address := comAddress(a)
	oldValue := db.MutableState.GetBalance(address)
	newValue, _ := common.CheckedU256Add(oldValue, v)
	db.MutableState.SetBalance(address, newValue)
}

func (db *database) GetBalance(a eth.Address) *big.Int {
	return db.MutableState.GetBalance(comAddress(a))
}

func (db *database) SetNonce(a eth.Address, v uint64) {
	db.MutableState.SetNonce(comAddress(a), v)
}

func (db *database) GetNonce(a eth.Address) uint64 {
	return db.MutableState.GetNonce(comAddress(a))
}

func (db *database) GetCodeHash(a eth.Address) eth.Hash {
	return ethHash(db.MutableState.GetCodeHash(comAddress(a)))
}

func (db *database) GetCodeSize(a eth.Address) int {
	return db.MutableState.GetCodeSize(comAddress(a))
}

func (db *database) GetCode(a eth.Address) []byte {
	return db.MutableState.GetCode(comAddress(a))
}

func (db *database) SetCode(a eth.Address, code []byte) {
	db.MutableState.SetCode(comAddress(a), code)
}

func (db *database) AddRefund(v *big.Int) {
	var err error
	db.Refund, err = libeth.CheckedU256Add(db.Refund, v)
	if err != nil {
		panic(err)
	}
}

func (db *database) GetRefund() *big.Int {
	return db.Refund
}

func (db *database) GetState(a eth.Address, k eth.Hash) eth.Hash {
	h, _ := db.MutableState.GetValue(comAddress(a), comHash(k))
	return ethHash(h)
}

func (db *database) SetState(a eth.Address, k eth.Hash, v eth.Hash) {
	db.MutableState.SetValue(comAddress(a), comHash(k), comHash(v))
}

func (db *database) Suicide(a eth.Address) bool {
	return db.MutableState.Suicide(comAddress(a))
}

func (db *database) HasSuicided(a eth.Address) bool {
	return db.MutableState.HasSuicide(comAddress(a))
}

// Exist reports whether the given account exists in state.
// Notably this should also return true for suicided accounts.
func (db *database) Exist(a eth.Address) bool {
	return db.MutableState.Exists(comAddress(a))
}

func (vm *nvm) RuleSet() ethvm.RuleSet { return vm.rules }
func (vm *nvm) Vm() ethvm.Vm           { return vm.evm }
func (vm *nvm) Origin() eth.Address    { return vm.origin }
func (vm *nvm) BlockNumber() *big.Int  { return vm.number }
func (vm *nvm) Coinbase() eth.Address  { return vm.coinbase }
func (vm *nvm) Time() *big.Int         { return vm.time }
func (vm *nvm) Difficulty() *big.Int   { return vm.difficulty }
func (vm *nvm) Db() ethvm.Database     { return vm.db }
func (vm *nvm) GasLimit() *big.Int     { return vm.gasLimit }
func (vm *nvm) VmType() ethvm.Type     { return ethvm.StdVmTy }

func (vm *nvm) GetHash(n uint64) eth.Hash {
	return ethHash(vm.blockhash(new(big.Int).SetUint64(n)))
}

func (vm *nvm) AddLog(log *ethvm.Log) {
	address := comAddress(log.Address)
	topics := make([]common.Hash, len(log.Topics))
	for i := 0; i < len(log.Topics); i++ {
		copy(topics[i][:], log.Topics[i][:])
	}
	vm.db.MutableState.AddLog(address, topics, log.Data)
}

func (vm *nvm) Depth() int     { return vm.depth }
func (vm *nvm) SetDepth(i int) { vm.depth = i }

func (vm *nvm) CanTransfer(from eth.Address, balance *big.Int) bool {
	if vm.skipTransfer {
		if vm.initial {
			vm.initial = false
			return true
		}
	}
	return vm.db.GetBalance(from).Cmp(balance) >= 0
}

func (vm *nvm) SnapshotDatabase() int {
	return int(vm.db.MutableState.Snapshot())
}

func (vm *nvm) RevertToSnapshot(snapshot int) {
	vm.db.MutableState.Revert(uint64(snapshot))
}

func (vm *nvm) Transfer(from, to ethvm.Account, amount *big.Int) {
	if vm.skipTransfer {
		return
	}
	ethc.Transfer(from, to, amount)
}

func (vm *nvm) Call(caller ethvm.ContractRef, addr eth.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	if vm.vmTest && vm.depth > 0 {
		caller.ReturnGas(gas, price)

		return nil, nil
	}
	ret, err := ethc.Call(vm, caller, addr, data, gas, price, value)
	vm.Gas = gas
	return ret, err
}

func (vm *nvm) CallCode(caller ethvm.ContractRef, addr eth.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	if vm.vmTest && vm.depth > 0 {
		caller.ReturnGas(gas, price)
		return nil, nil
	}
	return ethc.CallCode(vm, caller, addr, data, gas, price, value)
}

func (vm *nvm) DelegateCall(caller ethvm.ContractRef, addr eth.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	if vm.vmTest && vm.depth > 0 {
		caller.ReturnGas(gas, price)
		return nil, nil
	}
	return ethc.DelegateCall(vm, caller, addr, data, gas, price)
}

func (vm *nvm) Create(caller ethvm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, eth.Address, error) {
	var err error
	var ret []byte
	var addr eth.Address
	if vm.vmTest {
		address := comAddress(caller.Address())
		caller.ReturnGas(gas, price)
		nonce := vm.db.MutableState.GetNonce(address)
		obj := vm.db.GetOrNew(ethAddress(crypto.CreateAddress(address, nonce)))
		addr = obj.Address()
	} else {
		ret, addr, err = ethc.Create(vm, caller, data, gas, price, value)
	}
	return ret, addr, err
}
