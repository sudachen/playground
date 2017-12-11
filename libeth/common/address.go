package common

import (
	"math/big"
)

const AddressLength = 20

type Address [AddressLength]byte

func (a Address) Str() string   { return string(a[:]) }
func (a Address) Bytes() []byte { return a[:] }
func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }
func (a Address) Hash() Hash    { return BytesToHash(a[:]) }
func (a Address) Hex() string   { return "0x" + Bytes2Hex(a[:]) }

func (a *Address) SetString(s string) { a.SetBytes([]byte(s)) }

func StringToAddress(s string) Address { return BytesToAddress([]byte(s)) }
func BigToAddress(b *big.Int) Address  { return BytesToAddress(b.Bytes()) }
func HexToAddress(s string) Address    { return BytesToAddress(FromHex(s)) }

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a.Bytes()) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

func (a *Address) Set(other Address) {
	copy(a[:], other[:])
}

func (a Address) Cmp(b Address) int {
	for i := 0; i < AddressLength; i++ {
		c := int(a[i]) - int(b[i])
		if c != 0 {
			return c
		}
	}
	return 0
}

type SortableAdresses []Address

func (a SortableAdresses) Len() int           { return len(a) }
func (a SortableAdresses) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortableAdresses) Less(i, j int) bool { return a[i].Cmp(a[j]) < 0 }
