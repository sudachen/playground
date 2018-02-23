package ppftool

import (
	"fmt"

	"github.com/google/pprof/driver"
)

type ui struct {
	report *Report
}

func (u *ui) ReadLine(prompt string) (string, error) {
	return "quit", nil
}

func (u *ui) PrintErr(a ...interface{}) {
	if u.report != nil {
		u.report.Errors = append(u.report.Errors, fmt.Sprint(a...))
	}
}

func (u *ui) IsTerminal() bool { return false }

func (u *ui) Print(a ...interface{}) {}

func (u *ui) SetAutoComplete(complete func(string) string) {}

func FakeUi() driver.UI {
	return &ui{}
}
