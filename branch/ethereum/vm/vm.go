package vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/sudachen/playground/libeth"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/common"
)

type nvm struct {}

func NewVM() libeth.VM1 {
	return &nvm{}
}

func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

func Transfer(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}

func (n *nvm) Execute(
	msg core.Message,
	bi *libeth.BlockInfo,
	sdb vm.StateDB) (uint64, bool, error) {

	context := vm.Context{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:	 bi.Blockhash,
		Origin:      msg.From(),
		Coinbase:    bi.Coinbase,
		BlockNumber: new(big.Int).Set(bi.Number),
		Time:        new(big.Int).Set(bi.Time),
		Difficulty:  new(big.Int).Set(bi.Difficulty),
		GasLimit:    bi.GasLimit,
		GasPrice:    new(big.Int).Set(msg.GasPrice()),
	}

	gp := new(core.GasPool).AddGas(bi.GasLimit)
	e := vm.NewEVM(context, sdb, bi.Config, vm.Config{})
	// Apply the transaction to the current state (included in the env)
	_, gas, failed, err := core.ApplyMessage(e, msg, gp)
	return gas, failed, err
}
