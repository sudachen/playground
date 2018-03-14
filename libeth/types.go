package libeth

import (
	"github.com/ethereum/go-ethereum/common"
)

type Address = common.Address
type Hash = common.Hash

const AddressLength = common.AddressLength
const HashLength = common.HashLength

func CmpAddr(a Address, b Address) int {
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
func (a SortableAdresses) Less(i, j int) bool { return CmpAddr(a[i],a[j]) < 0 }
