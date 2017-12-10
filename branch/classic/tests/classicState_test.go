
package tests

import (
	"testing"
	"fmt"
	testutil "github.com/sudachen/playground/tests"
	"path/filepath"
	"github.com/sudachen/playground/libeth/common"
	"github.com/sudachen/playground/ethereum/crypto"
	"github.com/sudachen/playground/branch/classic/vm"
	"errors"
	"math/big"
)

var classicBaseDir      = filepath.Join("..", "..", "..", "testdata", "classic_test")
var classicStateTestDir = filepath.Join(classicBaseDir, "StateTests")

func TestStateClassic(t *testing.T) {
	//t.Parallel()

	st := new(testutil.TestMatcher)
	st.Walk(t, classicStateTestDir, func(t *testing.T, name string, test map[string]interface{}) {
		t.Run(name, func(t *testing.T) {
			err := RunTestClassic(t,test)
			st.CheckFailure(t, name, err)
		})
	})
}

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

func RunTestClassic(t *testing.T, test map[string]interface{}) error {

	var pre common.State
	var post common.State
	var tx  *common.Transaction
	var secretKey []byte
	var expectedOut []byte
	var err error

	fmt.Printf("== %s\n",t.Name())

	if pre, err = testutil.NewPreState(test); err != nil {return err}
	if post, err = testutil.NewClassicPostState(test); err != nil {return err}
	if tx, err = testutil.GetTransaction(test); err != nil {return err}
	if expectedOut, err = testutil.GetTransactionOut(test); err != nil {return err}
	if secretKey, err = testutil.GetSecretKey(test); err != nil {return err}

	tx.From = crypto.PubkeyToAddress(crypto.ToECDSA(secretKey).PublicKey)
	evm := vm.NewVM()

	blockInfo := &common.BlockInfo{
		Blockhash: func(n *big.Int)common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(n.String())))
		},
	}

	if err = testutil.FillBlockInfo(test,blockInfo); err != nil {return err}

	if out,_,state,err := evm.Execute(tx,blockInfo,pre); err != nil {
		return err
	} else {
		failed := false
		adrs := state.Compare(post)
		if len(adrs) != 0 {
			failed = true
		}
		if !equal(expectedOut,out) {
			failed = true
		}
		if failed {
			return errors.New("final state des not match to expected")
		}
	}

	return nil
}
