package classic

import (
	"math/big"

	"github.com/sudachen/benchmark"
	"github.com/sudachen/playground/crypto"
	"github.com/sudachen/playground/libeth"
	"github.com/ethereum/go-ethereum/common"
)

func StateBench(repeat int, test map[string]interface{}, name string, rules *libeth.RuleSet, evm libeth.VM, t *benchmark.T) error {
	var pre libeth.State
	var tx *libeth.Transaction
	var secretKey []byte
	var err error

	blockInfo := &libeth.BlockInfo{
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
