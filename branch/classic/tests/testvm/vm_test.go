package testvm

import (
	"github.com/sudachen/playground/branch/classic/vm"
	"github.com/sudachen/playground/playtool"
	"github.com/sudachen/playground/playtool/classic"
	"path/filepath"
	"testing"
)

var tfo = &playtool.Tfo{
	RootDir: filepath.Join("..", "..", "..", "..", "testdata", "classic_test", "StateTests"),
	NewVM:   vm.NewVM,
	Proc:    classic.StateTest,
}

func TestState(t *testing.T) {
	classic.RunAllStateTests(tfo,t)
}
