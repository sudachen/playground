package ppftool

import (
	"io/ioutil"
	"os"

	"github.com/google/pprof/driver"
)

func Top(b []byte, o *Options) (*Report, error) {
	tempfile := TempFileName()
	rpt := &Report{Unit: o.Unit}

	err := driver.PProf(&driver.Options{
		Fetch:   &fetcher{b},
		Flagset: o.flagset("-top", "-output="+tempfile),
		UI:      &ui{report: rpt},
	})

	if err != nil {
		return nil, err
	}

	if b, err := ioutil.ReadFile(tempfile); err != nil {
		return nil, err
	} else {
		rpt.WriteTop(b)
	}

	os.Remove(tempfile)

	return rpt, nil
}
