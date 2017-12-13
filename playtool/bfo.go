
package playtool

import (
	"github.com/sudachen/playground/benchmark"
	"github.com/sudachen/playground/libeth/common"
	"strings"
)

type Bfo struct {
	Proc    func(int,map[string]interface{},string,*common.RuleSet,common.VM,*benchmark.T)error
	NewVM   func() common.VM
	RootDir string
	Repeat  int
}

func (bfo *Bfo) RunAll(tests []*Nfo, t *benchmark.T) {
	for _, x := range tests {
		if !x.Pass {
			t.Run(x.Name, func(t0 *benchmark.T)error {

				t0.Pause()
				defer t0.Resume()

				return x.RunAllBenchmarks(bfo,t0)

			})
		}
	}
}

func (bfo *Bfo) RunOne(tests []*Nfo, name string, t *benchmark.T) {
	p := strings.Split(name,"/")
	t.Run(p[0], func(t0 *benchmark.T)error {

		t0.Pause()
		defer t0.Resume()

		nfo := FindTest(tests, p[0])
		if len(p) > 1 && p[1] != "*" {
			return nfo.RunOneBenchmark(bfo, p[1], t0)
		} else {
			return nfo.RunAllBenchmarks(bfo, t0)
		}

	})
}
