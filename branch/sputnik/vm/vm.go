package vm

import (
	etc "github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/sputnikvm-ffi/go/sputnikvm"
	"github.com/sudachen/playground/libeth/common"
	"github.com/sudachen/playground/libeth/state"
	"math/big"
	//etcc "github.com/ethereumproject/go-ethereum/core"
	//etcvm "github.com/ethereumproject/go-ethereum/core/vm"
)

var defaultRuleset = &common.RuleSet{
	HomesteadBlock:           big.NewInt(1150000),
	HomesteadGasRepriceBlock: big.NewInt(2500000),
	DiehardBlock:             big.NewInt(3000000),
	ExplosionBlock:           big.NewInt(5000000),
}

func comHash(a etc.Hash) (e common.Hash) {
	copy(e[:], a[:])
	return
}

func etcHash(a common.Hash) (e etc.Hash) {
	copy(e[:], a[:])
	return
}

func comAddress(a etc.Address) (e common.Address) {
	copy(e[:], a[:])
	return
}

func etcAddress(a common.Address) (e etc.Address) {
	copy(e[:], a[:])
	return
}

func etcAddressOpt(a *common.Address) *etc.Address {
	if a == nil {
		return nil
	}
	e := etcAddress(*a)
	return &e
}

type nvm struct{}

func NewVM() common.VM {
	return &nvm{}
}

func (*nvm) Execute(tx *common.Transaction, bi *common.BlockInfo, st common.State) (
	/*out*/ []byte,
	/*usedGas*/ *big.Int,
	/*resultState*/ common.State,
	/*executionError*/ error) {

	rs := state.NewMicroState(st)

	vmtx := sputnikvm.Transaction{
		Caller:   etcAddress(tx.From),
		GasPrice: tx.GasPrice,
		GasLimit: tx.GasLimit,
		Address:  etcAddressOpt(tx.To),
		Value:    tx.Value,
		Input:    tx.Data,
		Nonce:    new(big.Int).SetUint64(tx.Nonce),
	}

	vmheader := sputnikvm.HeaderParams{
		Beneficiary: etcAddress(bi.Coinbase),
		Timestamp:   bi.Time.Uint64(),
		Number:      bi.Number,
		Difficulty:  bi.Difficulty,
		GasLimit:    bi.GasLimit,
	}

	currentNumber := bi.Number

	rules := bi.RuleSet
	if rules == nil {
		rules = defaultRuleset
	}

	var vm *sputnikvm.VM
	if rules.DiehardBlock != nil && currentNumber.Cmp(rules.DiehardBlock) >= 0 {
		vm = sputnikvm.NewEIP160(&vmtx, &vmheader)
	} else if rules.HomesteadGasRepriceBlock != nil && currentNumber.Cmp(rules.HomesteadGasRepriceBlock) >= 0 {
		vm = sputnikvm.NewEIP150(&vmtx, &vmheader)
	} else if rules.HomesteadBlock != nil && currentNumber.Cmp(rules.HomesteadBlock) >= 0 {
		vm = sputnikvm.NewHomestead(&vmtx, &vmheader)
	} else {
		vm = sputnikvm.NewFrontier(&vmtx, &vmheader)
	}

Loop:
	for {
		ret := vm.Fire()
		switch ret.Typ() {
		case sputnikvm.RequireNone:
			break Loop
		case sputnikvm.RequireAccount:
			address := ret.Address()
			a := comAddress(address)
			if rs.Exists(a) {
				vm.CommitAccount(address, new(big.Int).SetUint64(rs.GetNonce(a)),
					rs.GetBalance(a), rs.GetCode(a))
				break
			}
			vm.CommitNonexist(address)
		case sputnikvm.RequireAccountCode:
			address := ret.Address()
			a := comAddress(address)
			if rs.Exists(a) {
				vm.CommitAccountCode(address, rs.GetCode(a))
				break
			}
			vm.CommitNonexist(address)
		case sputnikvm.RequireAccountStorage:
			address := ret.Address()
			a := comAddress(address)
			key := ret.StorageKey()
			if rs.Exists(a) {
				value, _ := rs.GetValue(a, common.BigToHash(key))
				vm.CommitAccountStorage(address, key, value.Big())
				break
			}
			vm.CommitNonexist(address)
		case sputnikvm.RequireBlockhash:
			number := ret.BlockNumber()
			hash := bi.Blockhash(number)
			vm.CommitBlockhash(number, etcHash(hash))
		}
	}

	// VM execution is finished at this point. We apply changes to the statedb.

	for _, account := range vm.AccountChanges() {
		switch account.Typ() {
		case sputnikvm.AccountChangeIncreaseBalance:
			address := account.Address()
			o := common.MutableAccount{rs, comAddress(address)}
			amount := account.ChangedAmount()
			o.AddBalance(amount)
		case sputnikvm.AccountChangeDecreaseBalance:
			address := account.Address()
			o := common.MutableAccount{rs, comAddress(address)}
			amount := account.ChangedAmount()
			o.SubBalance(amount)
		case sputnikvm.AccountChangeRemoved:
			address := account.Address()
			rs.Suicide(comAddress(address))
		case sputnikvm.AccountChangeFull, sputnikvm.AccountChangeCreate:
			address := account.Address()
			o := common.MutableAccount{rs, comAddress(address)}
			o.SetBalance(account.Balance())
			o.SetNonce(account.Nonce().Uint64())
			o.SetCode(account.Code())
			if account.Typ() == sputnikvm.AccountChangeFull {
				for _, item := range account.ChangedStorage() {
					o.SetValue(common.BigToHash(item.Key), common.BigToHash(item.Value))
				}
			} else {
				for _, item := range account.Storage() {
					o.SetValue(common.BigToHash(item.Key), common.BigToHash(item.Value))
				}
			}
		default:
			panic("unreachable")
		}
	}
	for _, log := range vm.Logs() {
		topics := make([]common.Hash, len(log.Topics))
		for i := 0; i < len(log.Topics); i++ {
			copy(topics[i][:], log.Topics[i][:])
		}
		rs.AddLog(comAddress(log.Address), topics, log.Data)
	}

	usedGas := vm.UsedGas()

	vm.Free()
	return nil, usedGas, rs.Freeze(), nil
}
