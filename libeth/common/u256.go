
package common

import (
	"math/big"
)

func CheckedU256Value(value *big.Int) (*big.Int, error) {
	return value, nil
}

func CheckedU256Add(value *big.Int, diff *big.Int) (*big.Int, error) {
	v := new(big.Int).Add(value,diff)
	return CheckedU256Value(v)
}


func CheckedU256Sub(value *big.Int, diff *big.Int) (*big.Int, error) {
	v := new(big.Int).Sub(value,diff)
	return CheckedU256Value(v)
}

