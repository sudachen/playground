package libeth

import (
	"math/big"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
)

type Transaction struct {
	Data     []byte
	GasLimit *big.Int
	GasPrice *big.Int
	Value    *big.Int
	Nonce    uint64
	To       *Address
	From     Address
}

type Message struct {
	*Transaction
}

type RuleSet struct {
	HomesteadBlock           *big.Int
	DAOForkBlock    		 *big.Int
	HomesteadGasRepriceBlock *big.Int
	DiehardBlock             *big.Int
	ExplosionBlock           *big.Int
}

type BlockInfo struct {
	types.Header
	BigGasLimit  *big.Int // Gas limit
	Blockhash    func(uint64) Hash
	RuleSet      *RuleSet
	Config       *params.ChainConfig
}

func (bi *BlockInfo) ResolveRules() *RuleSet {
	if bi.RuleSet != nil {
		return bi.RuleSet
	}
	hb := big.NewInt(1150000)
	if bi.Config != nil {
		hb = bi.Config.HomesteadBlock
	}
	dao := big.NewInt(1920000)
	if bi.Config != nil {
		dao = bi.Config.DAOForkBlock
	}
	return &RuleSet{
		HomesteadBlock:           hb,
		DAOForkBlock:             dao,
		HomesteadGasRepriceBlock: big.NewInt(2500000),
		DiehardBlock:             big.NewInt(3000000),
		ExplosionBlock:           big.NewInt(5000000),
	}
}

type VM interface {
	Execute(*Transaction, *BlockInfo, State) (
		out []byte,
		usedGas *big.Int,
		resultState State,
		executionError error)
}

type StateDB interface {
	vm.StateDB
	IntermediateRoot(deleteEmptyObjects bool) common.Hash
	GetLogs(hash common.Hash) []*types.Log
}

type VM1 interface {
	Execute(core.Message, *BlockInfo, vm.StateDB) (uint64, bool, error)
}
