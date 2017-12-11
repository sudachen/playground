package tests

import (
	"github.com/sudachen/playground/branch/classic/vm"
	testutil "github.com/sudachen/playground/tests"
	"path/filepath"
	"testing"
)

var tfo = &testutil.Tfo{
	RootDir: filepath.Join("..", "..", "..", "testdata", "classic_test", "StateTests"),
	NewVM:   vm.NewVM,
	Proc:    testutil.RunStateTests,
}

func TestState(t *testing.T) {
	testutil.RunClassicStateTests(t, tfo)
}
