package glog

import (
	"github.com/sudachen/misc/out"
)

func Errorf(t string, a ...interface{}) {
	out.Error.Printf(t,a...)
}

type Output bool

func V(p out.Level) Output { return Output(p.Visible()) }

func (Output) Infof(t string, a ...interface{}) {
	out.Info.Printf(t, a...)
}
