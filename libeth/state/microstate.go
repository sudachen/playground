package state

import (
	"github.com/sudachen/playground/libeth/common"
	"github.com/sudachen/playground/libeth/crypto"
	"math/big"
	"sort"
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

func (c Change) String() string {
	switch c {
	case None:
		return "None"
	case NoExists:
		return "NoExists"
	case Copyied:
		return "Copyied"
	case Created:
		return "Created"
	case Modified:
		return "Modified"
	case Suicide:
		return "Suicide"
	}
	return ""
}

type stCode struct {
	code []byte
	hash common.Hash
}

type stValue struct {
	changed bool
	value   common.Hash
}

type stData struct {
	snapshot uint64

	balance *big.Int
	nonce   uint64
	code    *stCode
	values  map[common.Hash]*stValue

	change     Change
	newborn    bool
	hasSuicide bool
}

type stAccount struct {
	data    *stData
	history []*stData
}

func (acc *stData) CopyFrom(origin common.State, address common.Address) {
	acc.nonce = origin.GetNonce(address)
	acc.hasSuicide = origin.HasSuicide(address)
	acc.newborn = false
	acc.balance = new(big.Int).Set(origin.GetBalance(address))
	codeHash := origin.GetCodeHash(address)
	if codeHash != (common.Hash{}) {
		originCode := origin.GetCode(address)
		acc.code = &stCode{make([]byte, len(originCode)), crypto.Keccak256Hash(originCode)}
		copy(acc.code.code, originCode)
	}
	origin.ProcessValues(address, func(key, val common.Hash) error {
		acc.values[key] = &stValue{false, val}
		return nil
	}, false)
}

/*var equalityError = errors.New("unequal")

func (acc *stData) IsEqualTo(st common.State, address common.Address) bool {
	hash := common.Hash{}
	if acc.code != nil {
		hash = acc.code.hash
	}
	if acc.balance.Cmp(st.GetBalance(address)) != 0 ||
		acc.nonce != st.GetNonce(address) ||
		acc.hasSuicide != st.HasSuicide(address) ||
		hash != st.GetCodeHash(address) {
		return false
	}
	mask := make(map[common.Hash]*stValue)
	for k, v := range acc.values {
		mask[k] = v
	}
	if nil != st.ProcessValues(address, func(key, val common.Hash) error {
		if v, exists := mask[key]; exists {
			delete(mask, key)
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
*/

type stLogs struct {
	records   []*common.Log
	snapshots map[uint64]int
}

type MicroState struct {
	state    map[common.Address]*stAccount
	snapshot uint64
	mutable  bool

	logs stLogs

	origin common.State
}

func NewMicroState(origin common.State) *MicroState {
	return &MicroState{
		state:    make(map[common.Address]*stAccount),
		logs:     stLogs{nil, make(map[uint64]int)},
		snapshot: 0,
		mutable:  true,
		origin:   origin,
	}
}

func (st *MicroState) Origin() common.State {
	return st.origin
}

func (st *MicroState) Exists(address common.Address) bool {
	if acc, exists := st.state[address]; exists {
		return acc.data.change != NoExists
	}
	if st.origin != nil {
		return st.origin.Exists(address)
	}
	return false
}

func (st *MicroState) HasSuicide(address common.Address) bool {
	if acc, exists := st.state[address]; exists {
		if acc.data.change != NoExists {
			return acc.data.hasSuicide
		}
	} else if st.origin != nil {
		return st.origin.HasSuicide(address)
	}
	return false
}

func (st *MicroState) GetBalance(address common.Address) *big.Int {
	if acc, exists := st.state[address]; exists {
		if acc.data.change != NoExists {
			return new(big.Int).Set(acc.data.balance)
		}
	} else if st.origin != nil {
		return st.origin.GetBalance(address)
	}
	return new(big.Int)
}

func (st *MicroState) GetNonce(address common.Address) uint64 {
	if acc, exists := st.state[address]; exists {
		if acc.data.change != NoExists {
			return acc.data.nonce
		}
	} else if st.origin != nil {
		return st.origin.GetNonce(address)
	}
	return 0
}

func (st *MicroState) GetCode(address common.Address) []byte {
	if acc, exists := st.state[address]; exists {
		if acc.data.change != NoExists && acc.data.code != nil {
			code := make([]byte, len(acc.data.code.code))
			copy(code, acc.data.code.code)
			return code
		}
	} else if st.origin != nil {
		return st.origin.GetCode(address)
	}
	return nil
}

func (st *MicroState) GetCodeHash(address common.Address) common.Hash {
	if acc, exists := st.state[address]; exists {
		if acc.data.change != NoExists && acc.data.code != nil {
			return acc.data.code.hash
		}
	} else if st.origin != nil {
		return st.origin.GetCodeHash(address)
	}
	return common.Hash{}
}

func (st *MicroState) GetCodeSize(address common.Address) int {
	if acc, exists := st.state[address]; exists {
		if acc.data.change != NoExists && acc.data.code != nil {
			return len(acc.data.code.code)
		}
	} else if st.origin != nil {
		return st.origin.GetCodeSize(address)
	}
	return 0
}

func (st *MicroState) GetValue(address common.Address, key common.Hash) (common.Hash, bool) {
	if acc, exists := st.state[address]; exists {
		if acc.data.change != NoExists {
			val, ok := acc.data.values[key]
			if ok {
				return val.value, true
			}
		}
	} else if st.origin != nil {
		return st.origin.GetValue(address, key)
	}
	return common.Hash{}, false
}

func (st *MicroState) ProcessValues(address common.Address, f func(common.Hash, common.Hash) error, changedOnly bool) error {
	if acc, exists := st.state[address]; exists {
		if acc.data.change != NoExists {
			for key, val := range acc.data.values {
				if val.changed || !changedOnly {
					if err := f(key, val.value); err != nil {
						return err
					}
				}
			}
		}
	} else if st.origin != nil && !changedOnly {
		return st.origin.ProcessValues(address, f, false)
	}
	return nil
}

func (st *MicroState) Addresses(changedOnly bool) []common.Address {
	var ret []common.Address
	var mask map[common.Address]bool

	if st.origin != nil && !changedOnly {
		ret = st.origin.Addresses(false)
		mask = make(map[common.Address]bool)
		for _, a := range ret {
			mask[a] = true
		}
	} else {
		ret = make([]common.Address, 0, len(st.state))
	}

	for a, acc := range st.state {
		if changedOnly || mask == nil || !mask[a] {
			if acc.data.change != NoExists &&
				(acc.data.change != Copyied || !changedOnly) {
				ret = append(ret, a)
			}
		}
	}

	sort.Sort(common.SortableAdresses(ret))
	return ret
}

func copyLogs(logs []*common.Log) []*common.Log {
	if logs != nil {
		ret := make([]*common.Log, len(logs))
		copy(ret, logs)
		return ret
	}
	return nil
}

func (st *MicroState) Immutable() common.State {
	if !st.mutable {
		return st
	}

	im := &MicroState{
		state:    make(map[common.Address]*stAccount),
		logs:     stLogs{copyLogs(st.logs.records), nil},
		snapshot: 0,
		mutable:  false,
		origin:   nil}

	for addr, dt := range st.state {
		acc := &stAccount{&stData{}, nil}
		*acc.data = *dt.data
		acc.data.snapshot = 0
		acc.data.values = make(map[common.Hash]*stValue)
		for k, v := range dt.data.values {
			acc.data.values[k] = v
		}
		im.state[addr] = acc
	}

	return im
}

func (st *MicroState) Freeze() common.State {
	st.mutable = false
	st.snapshot = 0
	for _, dt := range st.state {
		dt.history = nil
		dt.data.snapshot = 0
	}
	return st
}

func (st *MicroState) Create(address common.Address) bool {
	_, exists := st.state[address]
	st.Snap(address, Created)
	return !exists
}

func (st *MicroState) Snap(address common.Address, change Change) *stData {

	var acc *stAccount
	var exists bool

	if !st.mutable {
		panic("state is immutable!")
	}

	if acc, exists = st.state[address]; exists {

		if st.snapshot < acc.data.snapshot {
			panic("State corrupted: snapshot is invalid")
		} else if st.snapshot > acc.data.snapshot {
			old := acc.data
			acc.history = append(acc.history, old)
			acc.data = &stData{}
			if change != Created {
				*acc.data = *old
				acc.data.values = make(map[common.Hash]*stValue)
				for k, v := range old.values {
					acc.data.values[k] = v
				}
			} else {
				acc.data.values = make(map[common.Hash]*stValue)
				acc.data.balance = old.balance
			}
			acc.data.snapshot = st.snapshot
		}

		acc.data.change = change

	} else {

		if change == Suicide {
			if st.origin == nil || !st.origin.Exists(address) {
				return nil
			}
		}

		acc = &stAccount{&stData{
			snapshot: st.snapshot,
			values:   make(map[common.Hash]*stValue),
			balance:  new(big.Int),
			change:   change,
			newborn:  true,
		}, nil}

		if st.origin != nil {
			if change != Created {
				acc.data.CopyFrom(st.origin, address)
			} else {
				acc.data.balance = st.origin.GetBalance(address)
			}
		}

		st.state[address] = acc
	}

	if change == Suicide && acc.data.hasSuicide {
		return nil
	}

	return acc.data
}

func (st *MicroState) Suicide(address common.Address) bool {
	acc := st.Snap(address, Suicide)
	if acc != nil {
		acc.hasSuicide = true
		acc.balance = new(big.Int)
		return true
	}
	return false
}

func (st *MicroState) SetBalance(address common.Address, balance *big.Int) {
	acc := st.Snap(address, Modified)
	//fmt.Fprintf(os.Stderr,"set balance(%v) %v => %v\n",address.Hex(),acc.balance,balance)
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
	var hash common.Hash
	if len(code) != 0 {
		hash = crypto.Keccak256Hash(code)
	}
	acc.code = &stCode{make([]byte, len(code)), hash}
	copy(acc.code.code, code)
	return nil
}

func (st *MicroState) SetValue(address common.Address, key common.Hash, value common.Hash) {
	acc := st.Snap(address, Modified)
	acc.values[key] = &stValue{true, value}
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
		if acc.data.snapshot > snapshot {
			var prev *stData = nil
			for len(acc.history) > 0 {
				l := len(acc.history) - 1
				ss := acc.history[l]
				acc.history = acc.history[:l]
				if ss.snapshot <= snapshot {
					prev = ss
					break
				}
			}
			if prev != nil {
				acc.data = prev
			} else {
				delete(st.state, a)
			}
		}
	}

	if ln, ok := st.logs.snapshots[snapshot]; ok {
		if st.logs.records != nil {
			st.logs.records = st.logs.records[:ln]
		}
		for k := range st.logs.snapshots {
			if k >= snapshot {
				delete(st.logs.snapshots, k)
			}
		}
	}
}

func (st *MicroState) AddLog(address common.Address, topics []common.Hash, data []byte) {
	if topics == nil {
		topics = make([]common.Hash, 0)
	}
	if data == nil {
		data = make([]byte, 0)
	}
	st.logs.records = append(st.logs.records, &common.Log{
		Address: address,
		Topics:  topics,
		Data:    data,
	})
}

func (st *MicroState) Logs() []*common.Log {
	ret := make([]*common.Log, len(st.logs.records))
	for i, log := range st.logs.records {
		ret[i] = log.Clone()
	}
	return ret
}
