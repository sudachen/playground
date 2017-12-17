
package main

import (
	"path/filepath"
	"github.com/sudachen/playground/playtool"
	"github.com/sudachen/playground/playtool/classic"
	"github.com/sudachen/benchmark"
	"github.com/sudachen/playground/branch/classic/vm"
	"os"
)

var bfo = &playtool.Bfo{
	RootDir: filepath.Join("..", "..", "..", "..", "testdata", "classic_test", "StateTests"),
	NewVM:   vm.NewVM,
	Proc:    classic.StateBench,
	Repeat:  playtool.DefaultRepeat,
}

func main() {
	t := benchmark.Run(".",func(t *benchmark.T)error{
		classic.RunAllStateBenchmarks(bfo,t)
		//classic.RunOneStateBenchmark(bfo,"StateExample/*",t)
		return nil
	})
	t.WriteJson(os.Stdout)
}
