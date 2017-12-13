package classic

import (
	"github.com/sudachen/playground/libeth/common"
	"github.com/sudachen/playground/benchmark"
	"math/big"
	"errors"
	"github.com/sudachen/playground/libeth/state"
	"github.com/sudachen/playground/libeth/crypto"
)

func StateBench(repeat int, test map[string]interface{}, name string, rules *common.RuleSet, evm common.VM, t *benchmark.T) error {
	var pre common.State
	var post common.State
	var tx *common.Transaction
	var secretKey []byte
	var expectedOut []byte
	var err error

	t.Pause()
	defer t.Resume()

	if pre, err = NewPreState(test); err != nil {
		return err
	}
	if post, err = NewClassicPostState(test); err != nil {
		return err
	}
	if tx, err = GetTransaction(test); err != nil {
		return err
	}
	if expectedOut, err = GetTransactionOut(test); err != nil {
		return err
	}
	if secretKey, err = GetSecretKey(test); err != nil {
		return err
	}

	tx.From = crypto.PubkeyToAddress(crypto.ToECDSA(secretKey).PublicKey)

	blockInfo := &common.BlockInfo{
		Blockhash: func(n *big.Int) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(n.String())))
		},
		RuleSet: rules,
	}

	if err = FillBlockInfo(test, blockInfo); err != nil {
		return err
	}

	for i := 0; i <= repeat; i++ {
		t.Resume()
		out, _, st, _:= evm.Execute(tx, blockInfo, pre)
		t.Pause()

		itWasFailed := 0
		adrs := state.CompareWithoutSuicides(st, post)

		if len(adrs) != 0 {
			itWasFailed |= FailedByState
		}
		if out != nil && !equal(expectedOut, out) {
			itWasFailed |= FailedByRet
		}

		if itWasFailed != 0 {
			return errors.New("final state des not match to expected")
		}
	}

	return nil
}

func RunAllStateBenchmarks(bfo *Bfo, t *benchmark.T) {
	bfo.RunAll(StateTests,t)
}

func RunOneStateBenchmark(bfo *Bfo, name string, t *benchmark.T) {
	bfo.RunOne(StateTests,name,t)
}
