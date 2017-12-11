package tests

import (
	"errors"
	"fmt"
	"github.com/sudachen/playground/libeth/common"
	"github.com/sudachen/playground/libeth/crypto"
	"github.com/sudachen/playground/libeth/state"
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

func equal(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range b {
		if a[i] != x {
			return false
		}
	}
	return true
}

func RunStateTest(t *testing.T, test map[string]interface{}, name string, rules *common.RuleSet, evm common.VM) error {
	var pre common.State
	var post common.State
	var tx *common.Transaction
	var secretKey []byte
	var expectedOut []byte
	var err error

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

	out, _, st, err := evm.Execute(tx, blockInfo, pre)
	failed := false
	adrs := state.CompareWithoutSuicides(st, post)
	if len(adrs) != 0 {
		failed = true
		for _, a := range adrs {
			t.Error(state.DumpDiff(st, post, a))
		}
		t.Error("-- pre --")
		t.Error(state.Dump(pre))
		t.Error("-- result --")
		t.Error(state.Dump(st))
		t.Error("-- expected --")
		t.Error(state.Dump(post))
	}
	if !equal(expectedOut, out) {
		failed = true
	}
	if failed {
		return errors.New("final state des not match to expected")
	}

	return nil
}

func RunStateTests(t *testing.T, nfo *Nfo, tfo *Tfo) {

	skipNames := make(map[string]bool)
	for _, x := range nfo.Skip {
		skipNames[x] = true
	}

	path := filepath.Join(tfo.RootDir, nfo.File)
	var tests map[string]interface{}
	if err := ReadJsonFile(path, &tests); err != nil {
		t.Fatal(err)
	}
	keys := SortedMapKeys(tests)
	if nfo.SkipTo != common.NulStr {
		for len(keys) != 0 && keys[0] != nfo.SkipTo {
			keys = keys[1:]
		}
	}
	for _, k := range keys {
		if !skipNames[k] {
			oneTest := tests[k].(map[string]interface{})
			fmt.Fprintf(os.Stderr, "test %s/%s\n", nfo.Name, k)
			if err := RunStateTest(t, oneTest, k, nfo.Rules, tfo.NewVM()); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func RunClassicStateTests(t *testing.T, tfo *Tfo) {
	for _, x := range ClassicStateTests {
		if !x.Pass {
			if x.Proc != nil {
				x.Proc(t, x, tfo)
			} else {
				t.Run(x.Name, func(t *testing.T) {
					tfo.Proc(t, x, tfo)
				})
			}
		}
	}
}

var ClassicStateTests = []*Nfo{
	&Nfo{
		Pass:   false,
		Name:   "StateExample",
		File:   "stExample.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "SystemOperations",
		File:   "stSystemOperationsTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "PreCompiledContracts",
		File:   "stPreCompiledContracts.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "RecursiveCreate",
		File:   "stRecursiveCreate.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass: false,
		Name: "Special",
		File: "stSpecialTest.json",
		//Skip:   []string{"JUMPDEST_AttackwithJump", "OverflowGasMakeMoney", "StackDepthLimitSEC", "txCost-sec73"},
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "Refund",
		File:   "stRefundTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "BlockHash",
		File:   "stBlockHashTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass: true,
		Name: "InitCode",
		File: "stInitCodeTest.json",
		//Skip:   []string{"NotEnoughCashContractCreation", "OutOfGasContractCreation", "StackUnderFlowContractCreation", "TransactionCreateAutoSuicideContract", "TransactionCreateRandomInitCode", "TransactionCreateStopInInitcode"},
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "Logs",
		File:   "stLogTests.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "Transaction",
		File:   "stTransactionTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "Transition",
		File:   "stTransitionTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "CallCreateCallCode",
		File:   "stCallCreateCallCodeTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "CallCodes",
		File:   "stCallCodes.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "DelegateCall",
		File:   "stDelegatecallTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "Memory",
		File:   "stMemoryTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "MemoryStress",
		File:   "stMemoryStressTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "QuadraticComplexity",
		File:   "stQuadraticComplexityTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "Solidity",
		File:   "stSolidityTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "Wallet",
		File:   "stWalletTest.json",
		Skip:   []string{},
		SkipTo: common.NulStr,
		Rules:  &common.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "HomesteadSystemOperations",
		File:   filepath.Join("Homestead", "stSystemOperationsTest.json"),
		Skip:   []string{},
		SkipTo: common.NulStr,
	},
}
