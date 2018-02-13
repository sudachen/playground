//  +build !amd64

package ethash

func dagiFNV(r, p *node) {
	for w := 0; w < NodeWords; w++ {
		r.w[w] = r.w[w] * FnvPrime ^ p.w[w]
	}
}

