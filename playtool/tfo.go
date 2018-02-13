// Copyright 2017 (c) Alexey Sudachen
// Copyright 2015 The go-ethereum Authors
//
// Based on go-ethereum init_test.go
//

package playtool

import (
	"testing"
	"strings"
	"time"

	"github.com/sudachen/playground/libeth"
)

type Tfo struct {
	Proc    func(map[string]interface{},string,*libeth.RuleSet,libeth.VM,*testing.T)error
	NewVM   func() libeth.VM
	RootDir string
}

func (tfo *Tfo) RunAll(tests []*Nfo, t *testing.T) {
	for _, x := range tests {
		if !x.Pass {
			t.Run(x.Name, func(t0 *testing.T) {
				x.RunAll(tfo,t0)
			})
		}
	}
}

func (tfo *Tfo) RunOne(tests []*Nfo, name string, t *testing.T) {
	p := strings.Split(name,"/")
	t.Run("none", func(t *testing.T) {
		t.Parallel()
		time.Sleep(5 * time.Second)
	})
	t.Run(p[0], func(t *testing.T) {
		t.Parallel()
		nfo := FindTest(tests, p[0])
		if len(p) > 1 && p[1] != "*" {
			nfo.RunOne(tfo, p[1], t)
		} else {
			nfo.RunAll(tfo,t)
		}
	})
}
