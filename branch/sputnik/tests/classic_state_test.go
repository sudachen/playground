package tests

import (
	"github.com/sudachen/playground/branch/sputnik/vm"
	testutil "github.com/sudachen/playground/tests"
	"path/filepath"
	"testing"
	"time"
)

var tfo = &testutil.Tfo{
	RootDir: filepath.Join("..", "..", "..", "testdata", "classic_test", "StateTests"),
	NewVM:   vm.NewVM,
	Proc:    testutil.RunStateTests,
}

func TestState(t *testing.T) {
	testutil.RunClassicStateTests(t, tfo)
}

func TestStateSpecial(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		t.Parallel()
		time.Sleep(5 * time.Second)
	})
	t.Run("Special", func(t *testing.T) {
		t.Parallel()
		nfo := *testutil.FindTest(testutil.ClassicStateTests, "Special")
		nfo.SkipTo = "OverflowGasMakeMoney"
		testutil.RunStateTests(t, &nfo, tfo)
	})
}
