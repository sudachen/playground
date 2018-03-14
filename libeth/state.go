package libeth

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/common"
)

type State interface {
	Exists(Address) bool
	HasSuicide(Address) bool
	Origin() State

	// if account does not exist it is treated as empty account
	GetBalance(Address) *big.Int
	GetNonce(Address) uint64
	GetCode(Address) []byte
	GetCodeHash(Address) Hash
	GetCodeSize(Address) int
	GetValue(Address, Hash) (Hash, bool)
	ProcessValues(address Address, f func(Hash, Hash) error, changedOnly bool) error

	Immutable() State

	Addresses(changedOnly bool) []Address

	Logs() []*Log
}

var CodeRewriteError = errors.New("could not rewrite contract code")

type MutableState interface {
/*
	Exists(Address) bool
	HasSuicide(Address) bool
	Origin() State

	// if account does not exist it is treated as empty account
	GetBalance(Address) *big.Int
	GetNonce(Address) uint64
	GetCode(Address) []byte
	GetCodeHash(Address) Hash
	GetCodeSize(Address) int
	GetValue(Address, Hash) (Hash, bool)
	ProcessValues(address Address, f func(Hash, Hash) error, changedOnly bool) error

	Immutable() State

	Addresses(changedOnly bool) []Address

	Logs() []*Log
*/
	State
	Create(Address) bool
	Suicide(Address) bool

	// if account does not exist it will be created on following calls
	SetBalance(Address, *big.Int)
	SetNonce(Address, uint64)
	SetCode(Address, []byte) error
	SetValue(Address, Hash, Hash)

	AddLog(address Address, topics []Hash, data []byte)

	Snapshot() uint64
	Revert(uint64)
}

type MutableStateProxy struct {
	*state.StateDB
}

func (p *MutableStateProxy) Exists(a Address) bool {
	return p.StateDB.Exist(a)
}
func (p *MutableStateProxy) HasSuicide(a Address) bool {
	return p.HasSuicided(a)
}
func (p *MutableStateProxy) Origin() State {
	return nil
}
func (p *MutableStateProxy) ProcessValues(
	a Address, f func(Hash, Hash) error, changedOnly bool) error {
	p.ForEachStorage(a,func(k,v Hash)bool{ f(k,v); return true })
	return nil
}
func (p *MutableStateProxy) SetValue(a Address, k Hash, v Hash) {
	p.SetState(a,k,v)
}
func (p *MutableStateProxy) GetValue(a Address, k Hash) (Hash, bool) {
	v := p.GetState(a,k)
	return v, v != Hash{}
}
func (p *MutableStateProxy) SetCode(a Address, b []byte) error {
	p.StateDB.SetCode(a, b)
	return nil
}

func (p *MutableStateProxy) Immutable() State { return p }

func (p *MutableStateProxy) Addresses(changedOnly bool) []Address {
	dmp := p.RawDump()
	a := make([]Address,0,len(dmp.Accounts))
	for k := range dmp.Accounts {
		a = append(a, common.HexToAddress(k))
	}
	return a
}

func (p *MutableStateProxy) Logs() []*Log {
	return nil
}

func (p *MutableStateProxy) Create(a Address) bool {
	p.CreateAccount(a)
	return true
}

func (p *MutableStateProxy) AddLog(address Address, topics []Hash, data []byte) {
}

func (p *MutableStateProxy) Snapshot() uint64 {
	return uint64(p.StateDB.Snapshot())
}

func (p *MutableStateProxy) Revert(s uint64) {
	p.StateDB.RevertToSnapshot(int(s))
}
