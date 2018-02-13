package ethash

import (
	"fmt"

	"github.com/sudachen/playground/crypto"
	"github.com/ethereum/go-ethereum/common"

	ethh "github.com/ethereum/ethash"
)

type Ethash struct {
	*Light
	*ethh.Full
}

// New creates an instance of the proof of work.
func New() *Ethash {
	full := &ethh.Full{}
	full.Turbo(true)
	return &Ethash{&Light{}, full}
}

func NewForTesting() (*Ethash, error) {
	e, err := ethh.NewForTesting()
	if err != nil {
		return nil, err
	}
	return &Ethash{&Light{test: true}, e.Full }, nil
}

var sharedLight = &Light{}

// NewShared creates an instance of the proof of work., where a single instance
// of the Light cache is shared across all instances created with NewShared.
func NewShared() *Ethash {
	full := &ethh.Full{}
	full.Turbo(true)
	return &Ethash{sharedLight, full}
}

func GetSeedHash(blockNum uint64) ([]byte, error) {
	if blockNum >= epochLength*2048 {
		return nil, fmt.Errorf("block number too high, limit is %d", epochLength*2048)
	}
	sh := makeSeedHash(blockNum / epochLength)
	return sh[:], nil
}

func makeSeedHash(epoch uint64) (sh common.Hash) {
	for ; epoch > 0; epoch-- {
		sh = crypto.Sha3Hash(sh[:])
	}
	return sh
}
