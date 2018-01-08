package classic

import (
	"github.com/sudachen/playground/libeth/common"
	"github.com/sudachen/benchmark"
	"math/big"
	"github.com/sudachen/playground/libeth/crypto"
)

func StateBench(repeat int, test map[string]interface{}, name string, rules *common.RuleSet, evm common.VM, t *benchmark.T) error {
	var pre common.State
	var tx *common.Transaction
	var secretKey []byte
	var err error

	blockInfo := &common.BlockInfo{
		Blockhash: func(n *big.Int) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(n.String())))
		},
		RuleSet: rules,
	}

	if pre, err = NewPreState(test); err == nil {
		if tx, err = GetTransaction(test); err == nil {
			if secretKey, err = GetSecretKey(test); err == nil {
				if err = FillBlockInfo(test, blockInfo); err == nil {
					tx.From = crypto.PubkeyToAddress(crypto.ToECDSA(secretKey).PublicKey)
				}
			}
		}
	}

	if err != nil {
		return err
	}

	for i := 0; i <= repeat; i++ {
		t.Start()
		evm.Execute(tx, blockInfo, pre)
	}

	return nil
}

func RunAllStateBenchmarks(bfo *Bfo, t *benchmark.T) {
	bfo.RunAll(StateTests,t)
}

func RunOneStateBenchmark(bfo *Bfo, name string, t *benchmark.T) {
	bfo.RunOne(StateTests,name,t)
}
