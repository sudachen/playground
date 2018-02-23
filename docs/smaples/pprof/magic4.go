package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/sudachen/playground/misc/ppftool"
)

func main() {
	var bf bytes.Buffer

	pprof.StartCPUProfile(&bf)
	for s:= ""; len(s) < 100000;  {
		s = s + fmt.Sprintf("%d",len(s))
	}
	pprof.StopCPUProfile()

	rpt, err := ppftool.Top(bf.Bytes(), &ppftool.Options{Count: 5, Hide: []string{"runtime\\."}})

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	fmt.Printf("%10s %11s %s\n","flat","%flat","function")
	for _, row := range rpt.Rows {
		fmt.Printf("%10.3f %10.3f%% %s\n",row.Flat,row.FlatPercent,row.Function)
	}
}

