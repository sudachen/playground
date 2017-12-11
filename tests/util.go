// Copyright 2017 (c) Alexey Sudachen
// Copyright 2015 The go-ethereum Authors
//
// Based on go-ethereum init_test.go
//

package tests

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/playground/libeth/common"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"testing"
)

type Tfo struct {
	Proc    func(*testing.T, *Nfo, *Tfo)
	NewVM   func() common.VM
	RootDir string
}

type Nfo struct {
	Pass   bool
	Name   string
	File   string
	Skip   []string
	SkipTo string
	Rules  *common.RuleSet
	Proc   func(*testing.T, *Nfo, *Tfo)
}

func FindTest(tests []*Nfo, name string) *Nfo {
	for _, x := range tests {
		if x.Name == name {
			return x
		}
	}
	return nil
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
