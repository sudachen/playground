package classic

import (
	"errors"
	"math/big"
	"path/filepath"
	"testing"
	"bytes"
	"bufio"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sudachen/playground/crypto"
	"github.com/sudachen/playground/libeth"
	"github.com/sudachen/playground/libeth/state"
)

func StateTest(test map[string]interface{}, name string, rules *libeth.RuleSet, evm libeth.VM, t *testing.T) error {
	var pre libeth.State
	var post libeth.State
	var tx *libeth.Transaction
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

	blockInfo := &libeth.BlockInfo{
		Blockhash: func(n *big.Int) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(n.String())))
		},
		RuleSet: rules,
	}

	if err = FillBlockInfo(test, blockInfo); err != nil {
		return err
	}

	out, _, st, err := evm.Execute(tx, blockInfo, pre)

	itWasFailed := 0
	adrs := state.CompareWithoutSuicides(st, post)

	if len(adrs) != 0 {
		itWasFailed |= FailedByState
	}
	if out != nil && !equal(expectedOut, out) {
		itWasFailed |= FailedByRet
	}

	if itWasFailed != 0 {
		bf := new(bytes.Buffer)
		wr := bufio.NewWriter(bf)
		fmt.Fprintf(wr,"\n%s => execution result does not match to expected\n",name)
		if (itWasFailed & FailedByRet) != 0 {
			wr.WriteString("returned bad value\n")
			fmt.Fprintf(wr,"\treturned: %s\n",common.ToHex(out))
			fmt.Fprintf(wr,"\texpected: %s\n",common.ToHex(expectedOut))
		}
		if (itWasFailed & FailedByState) != 0 {
			for _, a := range adrs {
				state.WriteDiff(wr,st,post,a)
			}
			wr.WriteString("\n-- before --\n")
			state.WriteDump(wr,pre,"\t")
			wr.WriteString("\n-- after --\n")
			state.WriteDump(wr,st,"\t")
			wr.WriteString("\n-- expected --\n")
			state.WriteDump(wr,post,"\t")
			wr.WriteString("\n")
			wr.Flush()
		}
		t.Error(bf.String())
		return errors.New("final state des not match to expected")
	}

	return nil
}

func RunAllStateTests(tfo *Tfo, t *testing.T) {
	tfo.RunAll(StateTests,t)
}

func RunOneStateTest(tfo *Tfo, name string, t *testing.T) {
	tfo.RunOne(StateTests,name,t)
}

var StateTests = []*Nfo{
	&Nfo{
		Pass:   false,
		Name:   "StateExample",
		File:   "stExample.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "SystemOperations",
		File:   "stSystemOperationsTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "PreCompiledContracts",
		File:   "stPreCompiledContracts.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "RecursiveCreate",
		File:   "stRecursiveCreate.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass: false,
		Name: "Special",
		File: "stSpecialTest.json",
		//Skip:   []string{"JUMPDEST_AttackwithJump", "OverflowGasMakeMoney", "StackDepthLimitSEC", "txCost-sec73"},
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "Refund",
		File:   "stRefundTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "BlockHash",
		File:   "stBlockHashTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass: true,
		Name: "InitCode",
		File: "stInitCodeTest.json",
		//Skip:   []string{"NotEnoughCashContractCreation", "OutOfGasContractCreation", "StackUnderFlowContractCreation", "TransactionCreateAutoSuicideContract", "TransactionCreateRandomInitCode", "TransactionCreateStopInInitcode"},
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "Logs",
		File:   "stLogTests.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "Transaction",
		File:   "stTransactionTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "Transition",
		File:   "stTransitionTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "CallCreateCallCode",
		File:   "stCallCreateCallCodeTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "CallCodes",
		File:   "stCallCodes.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "DelegateCall",
		File:   "stDelegatecallTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "Memory",
		File:   "stMemoryTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "MemoryStress",
		File:   "stMemoryStressTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "QuadraticComplexity",
		File:   "stQuadraticComplexityTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "Solidity",
		File:   "stSolidityTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   false,
		Name:   "Wallet",
		File:   "stWalletTest.json",
		Skip:   []string{},
		SkipTo: libeth.NulStr,
		Rules:  &libeth.RuleSet{HomesteadBlock: big.NewInt(1000000)},
	},
	&Nfo{
		Pass:   true,
		Name:   "HomesteadSystemOperations",
		File:   filepath.Join("Homestead", "stSystemOperationsTest.json"),
		Skip:   []string{},
		SkipTo: libeth.NulStr,
	},
}
