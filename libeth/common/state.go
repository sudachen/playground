package common

import (
	"math/big"
	"errors"
	"fmt"
	"strings"
)

type Log struct {
	Address Address
	Topics  []Hash
	Data    []byte
}

func (log *Log) String() string {
	topics := make([]string,len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = t.Hex()
	}
	return fmt.Sprintf("Log{Address:%s, Topics:%s, Data:%s}",
		log.Address.Hex(),
		strings.Join(topics,","),
		Bytes2Hex(log.Data))
}

func (log *Log) Clone() *Log {
	topics := make([]Hash,len(log.Topics))
	copy(topics,log.Topics)
	data := make([]byte,len(log.Data))
	copy(data,log.Data)
	return &Log{log.Address,topics,data}
}

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
	GetValue(Address,Hash) (Hash,bool)
	ProcessValues(address Address, f func(Hash,Hash)error, changedOnly bool) error

	Immutable() State

	Addresses(changedOnly bool) []Address
	Compare(with State) []Address

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
	GetValue(Address,Hash) (Hash,bool)
	ProcessValues(address Address, f func(Hash,Hash)error, changedOnly bool) error

	Immutable() State

	Addresses(changedOnly bool) []Address
	Compare(with State) []Address

	Logs() []*Log

	Create(Address) bool
	Suicide(Address) bool

	// if account does not exist it will be created on following calls
	SetBalance(Address,*big.Int)
	SetNonce(Address, uint64)
	SetCode(Address, []byte) error
	SetValue(Address, Hash, Hash)

	AddLog(address Address, topics []Hash, data []byte)

	Snapshot() uint64
	Revert(uint64)
}
