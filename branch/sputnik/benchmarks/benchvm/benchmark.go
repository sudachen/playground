package main

import (
	"os"
	"path/filepath"

	"github.com/sudachen/benchmark"
	"github.com/sudachen/playground/branch/sputnik/vm"
	"github.com/sudachen/playground/playtool"
	"github.com/sudachen/playground/playtool/classic"

	// disable some tests
	_ "github.com/sudachen/playground/branch/sputnik/tests/_classic"
)

var bfo = &playtool.Bfo{
	RootDir: filepath.Join("..", "..", "..", "..", "testdata", "classic_test", "StateTests"),
	NewVM:   vm.NewVM,
	Proc:    classic.StateBench,
	Repeat:  playtool.DefaultRepeat, // run every test 10 times total
}

func main() {
	t := benchmark.Run(".", func(t *benchmark.T) error {
		classic.RunAllStateBenchmarks(bfo, t)
		return nil
	})
	t.WriteJson(os.Stdout)
}
