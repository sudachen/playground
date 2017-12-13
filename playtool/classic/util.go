
package classic

import "github.com/sudachen/playground/playtool"

type Nfo = playtool.Nfo
type Tfo = playtool.Tfo
type Bfo = playtool.Bfo

func equal(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range b {
		if a[i] != x {
			return false
		}
	}
	return true
}

const (
	FailedByState = 1
	FailedByRet   = 2
	FailedByError = 4
)
