package main

import (
	"bytes"
	"fmt"
	"runtime/pprof"

	"github.com/google/pprof/driver"
	"github.com/sudachen/playground/misc/ppftool"
)

func main() {
	var bf bytes.Buffer

	pprof.StartCPUProfile(&bf)
	for s := ""; len(s) < 100000; {
		s = s + fmt.Sprintf("%d", len(s))
	}
	pprof.StopCPUProfile()

	driver.PProf(&driver.Options{
		Fetch:   ppftool.Fetcher(bf.Bytes()),
		Flagset: ppftool.Flagset("-top", "-nodecount=5"),
	})
}
