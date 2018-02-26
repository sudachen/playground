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

	tempfile := "pprof.output.txt"
	unit := ppftool.DefaultUnit
	rpt := &ppftool.Report{Unit: unit}

	driver.PProf(&driver.Options{
		Fetch:   ppftool.Fetcher(bf.Bytes()),
		Flagset: ppftool.Flagset("-top", "-nodecount=5", "-unit="+unit.String(), "-output="+tempfile),
		UI:      ppftool.FakeUi(),
		Writer:  rpt,
	})

	/*if b, err := ioutil.ReadFile(tempfile); err == nil {
		rpt.Write(b)
		os.Remove(tempfile)
	}*/

	fmt.Printf("%10s %11s %s\n", "flat", "%flat", "function")
	for _, row := range rpt.Rows {
		fmt.Printf("%10.3f %10.3f%% %s\n", row.Flat, row.FlatPercent, row.Function)
	}
}
