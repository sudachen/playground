package testvm

import (
	"github.com/sudachen/playground/branch/sputnik/vm"
	"github.com/sudachen/playground/playtool"
	"github.com/sudachen/playground/playtool/classic"
	"path/filepath"
	"testing"

	// disable some tests
	_ "github.com/sudachen/playground/branch/sputnik/tests/_classic"
)

var tfo = &playtool.Tfo{
	RootDir: filepath.Join("..", "..", "..", "..", "testdata", "classic_test", "StateTests"),
	NewVM:   vm.NewVM,
	Proc:    classic.StateTest,
}

func TestAll(t *testing.T) {
	classic.RunAllStateTests(tfo, t)
}
