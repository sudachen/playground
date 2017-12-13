// Copyright 2017 (c) Alexey Sudachen

package playtool

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/playground/libeth/common"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"testing"
	"strings"
	"path/filepath"
	"github.com/sudachen/playground/benchmark"
)

type Nfo struct {
	Pass   bool
	Name   string
	File   string
	Skip   []string
	SkipTo string
	Rules  *common.RuleSet
}

func (nfo *Nfo) runAll(rootDir string,f func(string,map[string]interface{})error) error {
	skipNames := make(map[string]bool)
	for _, x := range nfo.Skip {
		skipNames[x] = true
	}
	path := filepath.Join(rootDir, nfo.File)
	var tests map[string]interface{}
	if err := ReadJsonFile(path, &tests); err != nil {
		return err
	}
	keys := SortedMapKeys(tests)
	if nfo.SkipTo != common.NulStr {
		for len(keys) != 0 && keys[0] != nfo.SkipTo {
			keys = keys[1:]
		}
	}
	for _, k := range keys {
		if !skipNames[k] {
			oneTest := tests[k].(map[string]interface{})
			if err := f(nfo.Name+"/"+k,oneTest); err != nil {
				return err
			}
		}
	}

	return nil
}

func (nfo *Nfo) RunAll(tfo *Tfo, t *testing.T) {
	nfo.runAll(tfo.RootDir,func(name string,test map[string]interface{})error{
		if err := tfo.Proc(test, name, nfo.Rules, tfo.NewVM(), t); err != nil {
			t.Error(err)
			return err
		}
		return nil
	})
}

func (nfo *Nfo) runOne(rootDir string,name string,f func(string,map[string]interface{})error) error {
	path := filepath.Join(rootDir, nfo.File)

	var tests map[string]interface{}
	if err := ReadJsonFile(path, &tests); err != nil {
		return err
	}

	if m,ok := tests[name]; ok {
		oneTest := m.(map[string]interface{})
		if err := f(name,oneTest); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("test %s/%s does not exist",nfo.Name,name)
	}

	return nil
}

func (nfo *Nfo) RunOne(tfo *Tfo, name string, t *testing.T) {
	nfo.runOne(tfo.RootDir,name,func(name string,test map[string]interface{})error {
		if err := tfo.Proc(test, name, nfo.Rules, tfo.NewVM(), t); err != nil {
			t.Error(err)
			return err
		}
		return nil
	})
}

func (nfo *Nfo) getRunnbale(bfo *Bfo,t *benchmark.T) func(name string,test map[string]interface{})error {
	return func(name string,test map[string]interface{})error {
		//fmt.Printf("%s\n",name)
		t.Resume()
		defer t.Pause()

		return t.Run(name,func(t0 *benchmark.T)error {
			if err := bfo.Proc(bfo.Repeat,test, name, nfo.Rules, bfo.NewVM(), t0); err != nil {
				t.Error(err)
				return err
			}
			return nil
		})
	}
}

func (nfo *Nfo) RunOneBenchmark(bfo *Bfo, name string, t *benchmark.T) error {
	return nfo.runOne(bfo.RootDir,name,nfo.getRunnbale(bfo,t))
}

func (nfo *Nfo) RunAllBenchmarks(bfo *Bfo, t *benchmark.T) error {
	return nfo.runAll(bfo.RootDir,nfo.getRunnbale(bfo,t))
}

func FindTest(tests []*Nfo, name string) *Nfo {
	for _, x := range tests {
		if x.Name == name {
			return x
		}
	}
	return nil
}

func isIn(val string, a []string) bool {
	for _, x := range a {
		if x == val {
			return true
		}
	}
	return false
}

func SkipTests(tests []*Nfo, names ...string) {
	for _, x := range names {
		p := strings.Split(x,"/")
		if nfo := FindTest(tests,p[0]); nfo == nil {
			panic("there is no test group "+p[0])
		} else if p[1] == "*" {
			nfo.Pass = true
		} else {
			if !isIn(p[1],nfo.Skip) {
				nfo.Skip = append(nfo.Skip,p[1])
			}
		}
	}
}

func ReadJsonFile(fn string, value interface{}) error {
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer file.Close()

	err = readJson(file, value)
	if err != nil {
		return fmt.Errorf("%s in file %s", err.Error(), fn)
	}
	return nil
}

func readJson(reader io.Reader, value interface{}) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %v", err)
	}
	if err = json.Unmarshal(data, &value); err != nil {
		if syntaxerr, ok := err.(*json.SyntaxError); ok {
			line := findLine(data, syntaxerr.Offset)
			return fmt.Errorf("JSON syntax error at line %v: %v", line, err)
		}
		return err
	}
	return nil
}

func findLine(data []byte, offset int64) (line int) {
	line = 1
	for i, r := range string(data) {
		if int64(i) >= offset {
			return
		}
		if r == '\n' {
			line++
		}
	}
	return
}

func SortedMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
