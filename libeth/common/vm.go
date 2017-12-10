
package common

import "math/big"

type Transaction struct {
	Data []byte

	GasLimit *big.Int
	GasPrice *big.Int
	Value 	 *big.Int

	Nonce uint64

	To   *Address
	From  Address
}

type BlockInfo struct {
	Coinbase     Address
	ParentHash   Hash

	Number      *big.Int       // Block number
	Difficulty  *big.Int       // Difficulty for the current block
	GasLimit    *big.Int       // Gas limit
	Time        *big.Int       // Creation time

	Blockhash    func(*big.Int)Hash
}

type VM interface {
	Execute(*Transaction,*BlockInfo,State)(
		out []byte,
		usedGas *big.Int,
		resultState State,
		executionError error)
}

