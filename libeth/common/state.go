package common

import (
	"errors"
	"math/big"
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
