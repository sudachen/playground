// +build amd64

package ethash

// This function is implemented in dagi_amd64.s.

//go:noescape

func dagiFNV(r, p *node)
