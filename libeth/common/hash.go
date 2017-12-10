
package common

import (
	"math/big"
	"reflect"
	"math/rand"
)

const HashLength = 32;
type Hash [HashLength]byte;

func (h Hash) Str() string   { return string(h[:]) }
func (h Hash) Bytes() []byte { return h[:] }
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }
func (h Hash) Hex() string   { return "0x" + Bytes2Hex(h[:]) }
func (h Hash) IsEmpty() bool { return h == Hash{} }

func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }
func HexToHash(s string) Hash   { return BytesToHash(FromHex(s)) }

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h.Bytes()) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

func (h *Hash) SetString(s string) { h.SetBytes([]byte(s)) }

func (h *Hash) Set(other Hash) {
	for i, v := range other {
		h[i] = v
	}
}

func (h Hash) Generate(rand *rand.Rand, size int) reflect.Value {
	m := rand.Intn(len(h))
	for i := len(h) - 1; i > m; i-- {
		h[i] = byte(rand.Uint32())
	}
	return reflect.ValueOf(h)
}
