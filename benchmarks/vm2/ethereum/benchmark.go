package main

import (
	"github.com/sudachen/benchmark"
	"github.com/sudachen/playground/branch/ethereum/chain"
	et "github.com/sudachen/playground/playtool/ethereum"
)

const BatchLength = 2000
var opt = &chain.Options{
	ExportQueLen: 100,
}

func main() {
	bm := benchmark.Run(".", func(t *benchmark.T) error {
		return et.ChainBench(opt, BatchLength, 0, t, nil)
	})
	bm.WriteJsonResult()
}
