
package state

import (
	"github.com/sudachen/playground/libeth/common"
	"math/big"
	"errors"
	"github.com/sudachen/playground/ethereum/crypto"
)

type Change byte

const (
	None Change = iota
	NoExists
	Copyied
	Created
	Modified
	Suicide
)

type stCode struct {
	code []byte
	hash common.Hash
}

type stValue struct {
	changed bool
	value common.Hash
}

type stData struct {
	snapshot uint64

	balance  *big.Int
	nonce    uint64
	code     *stCode
	values   map[common.Hash]*stValue

	change   Change
	newborn  bool
	hasSuicide bool
}

type stAccount struct {
	data *stData
	history []*stData
}

func (acc *stData) CopyFrom(origin common.State, address common.Address) {
	acc.balance = new(big.Int).Set(origin.GetBalance(address))
	acc.nonce = origin.GetNonce(address)
	acc.hasSuicide = origin.HasSuicide(address)
	codeHash := origin.GetCodeHash(address)
	if codeHash != (common.Hash{}) {
		originCode := origin.GetCode(address)
		acc.code = &stCode{make([]byte,len(originCode)),crypto.Keccak256Hash(originCode) }
		copy(acc.code.code,originCode)
	}
	origin.ProcessValues(address, func( key, val common.Hash)error {
		acc.values[key] = &stValue{false,val}
		return nil
	}, false)
}

var equalityError = errors.New("unequal")

func (acc *stData) IsEqualTo(st common.State, address common.Address) bool {
	hash := common.Hash{}
	if acc.code != nil {
		hash = acc.code.hash
	}
	if  acc.balance.Cmp(st.GetBalance(address)) != 0 ||
		acc.nonce != st.GetNonce(address) ||
		acc.hasSuicide != st.HasSuicide(address) ||
		hash != st.GetCodeHash(address) {
		return false
	}
	mask := make(map[common.Hash]*stValue)
	for k,v := range acc.values {
		mask[k] = v
	}
	if nil != st.ProcessValues(address, func( key, val common.Hash)error {
		if v,exists := mask[key]; exists {
			delete(mask,key)
			if v.value != val {
				return equalityError
			}
		} else {
			return equalityError
		}
		return nil
	}, false) {
		return false
	}

	if len(mask) != 0 {
		return false
	}

	return true
}

type stLogs struct {
	records []*common.Log
	snapshots map[uint64]int
}

type MicroState struct {

	state map[common.Address]*stAccount
	snapshot uint64
	mutable bool

	logs stLogs

	origin common.State
}

func NewMicroState(origin common.State) *MicroState {
	return &MicroState{
		state: make(map[common.Address]*stAccount),
		logs: stLogs{nil,make(map[uint64]int)},
		snapshot: 0,
		mutable: true,
		origin: origin,
	}
}

func (st *MicroState) Origin() common.State {
	return st.origin
}

func (st *MicroState) QueryOrigin(address common.Address) *stData {
	var acc *stData = nil
	if st.mutable && st.origin != nil {
		if st.origin.Exists(address) {
			acc = st.Snap(address,Copyied)
			acc.CopyFrom(st.origin, address)
		} else {
			acc = st.Snap(address,NoExists)
		}
	}
	return acc
}

func (st *MicroState) Query(address common.Address) *stData {

	if acc, exists := st.state[address]; exists {
		return acc.data
	}

	if st.mutable && st.origin != nil {
		return st.QueryOrigin(address)
	}

	return nil
}

func (st *MicroState) Exists(address common.Address) bool {
	if acc := st.Query(address); acc != nil && acc.change != NoExists {
		return true
	}
	return false
}

func (st *MicroState) HasSuicide(address common.Address) bool {
	if acc := st.Query(address); acc != nil && acc.hasSuicide {
		return true
	}
	return false
}

func (st *MicroState) GetBalance(address common.Address) *big.Int {
	if acc := st.Query(address); acc != nil {
		return new(big.Int).Set(acc.balance)
	}
	return new(big.Int)
}

func (st *MicroState) GetNonce(address common.Address) uint64 {
	if acc := st.Query(address); acc != nil {
		return acc.nonce
	}
	return 0
}

func (st *MicroState) GetCode(address common.Address) []byte {
	if acc := st.Query(address); acc != nil && acc.code != nil {
		code := make([]byte,len(acc.code.code))
		copy(code,acc.code.code)
		return code
	}
	return nil
}

func (st *MicroState) GetCodeHash(address common.Address) common.Hash {
	if acc := st.Query(address); acc != nil && acc.code != nil {
		return acc.code.hash
	}
	return common.Hash{}
}

func (st *MicroState) GetCodeSize(address common.Address) int {
	if acc := st.Query(address); acc != nil && acc.code != nil {
		return len(acc.code.code)
	}
	return 0
}

func (st *MicroState) GetValue(address common.Address,key common.Hash) (common.Hash,bool) {
	if acc := st.Query(address); acc != nil && acc.change != NoExists {
		val, ok := acc.values[key]
		return val.value, ok
	}
	return common.Hash{}, false
}

func (st *MicroState) ProcessValues(address common.Address, f func (common.Hash,common.Hash) error, changedOnly bool) error {
	if acc := st.Query(address); acc != nil && acc.change != NoExists {
		for key,val := range acc.values {
			if val.changed || !changedOnly {
				if err := f(key, val.value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (st *MicroState) Addresses(changedOnly bool) []common.Address {
	ret := make([]common.Address,0,len(st.state))
	for a,acc := range st.state {
		if  acc.data.change != NoExists &&
			(acc.data.change != Copyied || !changedOnly) {
			ret = append(ret, a)
		}
	}
	return ret
}

func (st *MicroState) Compare(with common.State) []common.Address {

	if with == st.Origin() {
		return st.Addresses(true)
	}

	ret := make([]common.Address,0,len(st.state))

	other := make(map[common.Address]bool)
	for _, a := range with.Addresses(false) {
		other[a] = true
	}

	for a, acc := range st.state {
		if _, exists := other[a]; exists {
			other[a] = false
			if !acc.data.IsEqualTo(with,a) {
				ret = append(ret,a)
			}
		} else {
			ret = append(ret,a)
		}
	}

	for a,ok := range other {
		if ok {
			ret = append(ret,a)
		}
	}

	return ret
}

func copyLogs(logs []*common.Log) []*common.Log {
	if logs != nil {
		ret := make([]*common.Log,len(logs))
		copy(ret,logs)
		return ret
	}
	return nil
}

func (st *MicroState) Immutable() common.State {
	if !st.mutable {
		return st
	}

	im := &MicroState{
		state:make(map[common.Address]*stAccount),
		logs: stLogs{copyLogs(st.logs.records),nil},
		snapshot: 0,
		mutable: false,
		origin: nil }

	for addr, dt := range st.state {
		acc := &stAccount{&stData{},nil}
		*acc.data = *dt.data
		acc.data.values = make(map[common.Hash]*stValue)
		for k,v := range dt.data.values {
			acc.data.values[k] = v
		}
		im.state[addr] = acc
	}

	return im
}

func (st *MicroState) Freeze() common.State {
	st.mutable = false
	return st
}

func (st *MicroState) Create(address common.Address) bool {
	if _, exists := st.state[address]; exists {
		return false
	}

	acc := &stAccount{&stData{
		snapshot: st.snapshot,
		values: make(map[common.Hash]*stValue),
		balance: new(big.Int),
		change: Created,
		newborn: true,
	}, nil}

	st.state[address] = acc
	return true
}

func (st *MicroState) Snap(address common.Address, change Change) *stData {

	if !st.mutable {
		panic("state is immutable!")
	}

	if acc, exists := st.state[address]; exists {

		if st.snapshot < acc.data.snapshot {
			panic("State corrupted: snapshot is invalid")
		} else {
			if st.snapshot != acc.data.snapshot {
				old := acc.data
				acc.history = append(acc.history, old)
				acc.data = &stData{}
				if change != Suicide {
					*acc.data = *old
					acc.data.values = make(map[common.Hash]*stValue)
					for k,v := range old.values {
						acc.data.values[k] = v
					}
				} else {
					acc.data.nonce = old.nonce
				}
			}
		}

		acc.data.change = change
		return acc.data

	} else {

		acc := &stAccount{&stData{
			snapshot: st.snapshot,
			values: make(map[common.Hash]*stValue),
			balance: new(big.Int),
			change: change,
			newborn: true,
		}, nil}

		if st.origin != nil {
			acc.data.CopyFrom(st.origin,address)
		}

		st.state[address] = acc

		return acc.data
	}
}

func (st *MicroState) Suicide(address common.Address) bool {

	if st.Exists(address) {
		st.Snap(address, Suicide)
		return true
	}

	return false
}

func (st *MicroState) SetBalance(address common.Address,balance *big.Int) {
	acc := st.Snap(address, Modified)
	acc.balance = new(big.Int).Set(balance)
}

func (st *MicroState) SetNonce(address common.Address, nonce uint64) {
	acc := st.Snap(address, Modified)
	acc.nonce = nonce
}

func (st *MicroState) SetCode(address common.Address, code []byte) error {
	acc := st.Snap(address, Modified)
	if acc.code != nil {
		return common.CodeRewriteError
	}
	acc.code = &stCode{ make([]byte,len(code)), crypto.Keccak256Hash(code) }
	copy(acc.code.code,code)
	return nil
}

func (st *MicroState) SetValue(address common.Address, key common.Hash, value common.Hash) {
	acc := st.Snap(address, Modified)
	acc.values[key] = &stValue{true,value }
}

func (st *MicroState) Snapshot() uint64 {
	s := st.snapshot
	st.logs.snapshots[st.snapshot] = len(st.logs.records)
	st.snapshot++
	return s
}

func (st *MicroState) Revert(snapshot uint64) {

	if snapshot >= st.snapshot {
		panic("impossible snapshot number to revert")
	}

	for a, acc := range st.state {
		var prev *stData = nil
		for len(acc.history) > 0 {
			l := len(acc.history) - 1
			ss := acc.history[l]
			acc.history = acc.history[:l]
			if ss.snapshot <= snapshot {
				prev = ss
			}
		}
		if prev != nil {
			acc.data = prev
		} else {
			delete(st.state,a)
		}
	}

	if ln, ok := st.logs.snapshots[snapshot]; ok {
		if st.logs.records != nil {
			st.logs.records = st.logs.records[:ln]
		}
		for k := range st.logs.snapshots {
			if k >= snapshot {
				delete(st.logs.snapshots,k)
			}
		}
	}
}

func (st *MicroState) AddLog(address common.Address, topics []common.Hash, data []byte) {
	if topics == nil { topics = make([]common.Hash,0)}
	if data == nil { data = make([]byte,0)}
	st.logs.records = append(st.logs.records,&common.Log{
		Address: address,
		Topics: topics,
		Data: data,
	})
}

func (st *MicroState) Logs() []*common.Log {
	ret := make([]*common.Log,len(st.logs.records))
	for i, log := range st.logs.records {
		ret[i] = log.Clone()
	}
	return ret
}
