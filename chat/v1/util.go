package v1

import "time"

func done(q chan struct{}) bool {
	select {
	case <-q:
		return true
	default:
		return false
	}
}

func done2(q chan struct{}, t <-chan time.Time) bool {
	select {
	case <-q:
		return true
	case <-t:
		return false
	}
}
