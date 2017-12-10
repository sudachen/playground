
package common

import (
	"math/big"
	"fmt"
	"errors"
)

type Account struct {
	State State
	Address Address
}

func (a Account) String() string {
	return fmt.Sprintf("Account{%s}",a.Address.Hex())
}

func (a *Account) Balance() *big.Int { return a.State.GetBalance(a.Address) }
func (a *Account) Nonce() uint64 { return a.State.GetNonce(a.Address) }
func (a *Account) Exists() bool { return a.State.Exists(a.Address) }
func (a *Account) HasSuicide() bool { return a.State.HasSuicide(a.Address) }
func (a *Account) Code() []byte { return a.State.GetCode(a.Address) }
func (a *Account) CodeHash() Hash { return a.State.GetCodeHash(a.Address) }
func (a *Account) CodeSize() int { return a.State.GetCodeSize(a.Address) }
func (a *Account) Immutable() *Account { return a }

func (a *Account) GetValue(key Hash) (Hash,bool) {
	return a.State.GetValue(a.Address,key)
}

func (a *Account) ProcessValues(f func(Hash,Hash)error, changedOnly bool) error {
	return a.State.ProcessValues(a.Address,f,changedOnly)
}

func NewAccount(address Address, state State) *Account { return &Account{ state, address } }

type MutableAccount struct {
	State MutableState
	Address Address
}

func (a *MutableAccount) Balance() *big.Int { return a.State.GetBalance(a.Address) }
func (a *MutableAccount) Nonce() uint64     { return a.State.GetNonce(a.Address) }
func (a *MutableAccount) Exists() bool      { return a.State.Exists(a.Address) }
func (a *MutableAccount) HasSuicide() bool  { return a.State.HasSuicide(a.Address) }
func (a *MutableAccount) Code() []byte      { return a.State.GetCode(a.Address) }
func (a *MutableAccount) CodeHash() Hash    { return a.State.GetCodeHash(a.Address) }
func (a *MutableAccount) CodeSize() int     { return a.State.GetCodeSize(a.Address) }

func (a *MutableAccount) SetBalance(balance *big.Int) error {
	if checkedValue, err := CheckedU256Value(balance); err != nil {
		return NewAccountError(a.Immutable(),err)
	} else {
		a.State.SetBalance(a.Address, checkedValue)
		return nil
	}
}

func (a *MutableAccount) AddBalance(diff *big.Int) error {
	balance := a.State.GetBalance(a.Address)
	if newBalance, err := CheckedU256Add(balance,diff); err != nil {
		return NewAccountError(a.Immutable(),err)
	} else {
		a.State.SetBalance(a.Address, newBalance)
		return nil
	}
}

func (a *MutableAccount) SubBalance(diff *big.Int) error {
	balance := a.State.GetBalance(a.Address)
	if newBalance, err := CheckedU256Sub(balance,diff); err != nil {
		return NewAccountError(a.Immutable(),err)
	} else {
		a.State.SetBalance(a.Address, newBalance)
		return nil
	}
}

func (a *MutableAccount) SetNonce(nonce uint64) { a.State.SetNonce(a.Address,nonce) }
func (a *MutableAccount) SetCode(code []byte) error { return a.State.SetCode(a.Address,code) }
func (a *MutableAccount) SetValue(key Hash, value Hash) { a.State.SetValue(a.Address,key,value)}
func (a *MutableAccount) Suicide() bool { return a.State.Suicide(a.Address) }

func (a *MutableAccount) GetValue(key Hash) (Hash,bool) {
	return a.State.GetValue(a.Address,key)
}

func (a *MutableAccount) ProcessValues(f func(Hash,Hash)error, changedOnly bool) error {
	return a.State.ProcessValues(a.Address,f,changedOnly)
}

func NewMutableAccount(address Address, state MutableState) *MutableAccount { return &MutableAccount{state, address}}
func (a *MutableAccount) Immutable() *Account { return &Account{a.State.Immutable(), a.Address} }

func CreateNewAccount(address Address, balance *big.Int, nonce uint64, state MutableState) *MutableAccount {
	if state.Create(address) {
		state.SetBalance(address,balance)
		state.SetNonce(address,nonce)
		return &MutableAccount{state, address}
	}
	return nil
}

type AccountError struct {
	Account
	Reason error
}

func (e *AccountError) Error() string {
	return fmt.Sprintf("%v: %v",e.Account,e.Reason)
}

func NewAccountError(a *Account, err interface{}) error {
	switch err.(type) {
	case error:
		return &AccountError{*a,err.(error)}
	case string:
		return &AccountError{*a,errors.New(err.(string))}
	default:
		return &AccountError{*a,fmt.Errorf("%s",err)}
	}
}
