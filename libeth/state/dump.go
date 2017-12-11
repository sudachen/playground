package state

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/sudachen/playground/libeth/common"
)

func Dump(st common.State) string {
	var bf bytes.Buffer
	wr := bufio.NewWriter(&bf)
	addresses := st.Addresses(false)
	fmt.Fprintf(wr, "immutable state: %d accounts\n", len(addresses))
	for i, a := range addresses {
		o := common.Account{st, a}
		fmt.Fprintf(wr, "%d\tAddress: %v\n", i, o.Address.Hex())
		fmt.Fprintf(wr, "\tBalance: %v\n", o.Balance())
		fmt.Fprintf(wr, "\tNonce: %v\n", o.Nonce())
		fmt.Fprintf(wr, "\tSuicided: %v\n", o.HasSuicide())
		fmt.Fprintf(wr, "\tCode: %v\n", common.ToHex(o.Code()))
		fmt.Fprintf(wr, "\tHash: %v\n", o.CodeHash().Hex())
		o.ProcessValues(func(k, v common.Hash) error {
			fmt.Fprintf(wr, "\t\t%v => %v\n", k.Hex(), v.Hex())
			return nil
		}, false)
		wr.Flush()
	}
	return bf.String()
}

func DumpDiff(st common.State, etalonSt common.State, address common.Address) string {
	o := common.Account{st, address}
	e := common.Account{etalonSt, address}
	if o.Exists() != e.Exists() {
		if !o.HasSuicide() && !e.HasSuicide() {
			var oExists = "exists"
			var eExists = "it have not"
			if !o.Exists() {
				oExists = "does not exist"
				eExists = "it have to be"
			}
			return fmt.Sprintf("address %s %s but %s",
				address.Hex(), oExists, eExists)
		}
	} else {
		r := ""
		if o.Balance().Cmp(e.Balance()) != 0 {
			r = r + fmt.Sprintf("balance of %s is %v but have to be %v\n",
				address.Hex(), o.Balance(), e.Balance())
		}
		if o.Nonce() != e.Nonce() {
			r = r + fmt.Sprintf("nonce of %s is %v but have to be %v\n",
				address.Hex(), o.Nonce(), e.Nonce())
		}
		return r
	}
	return ""
}

func Compare(sta common.State, stb common.State) []common.Address {
	staa := sta.Addresses(false)
	ret := make([]common.Address, 0, len(staa))

	other := make(map[common.Address]bool)
	for _, a := range staa {
		other[a] = true
	}

	for _, a := range stb.Addresses(false) {
		if _, exists := other[a]; exists {
			other[a] = false
			oa := &common.Account{sta, a}
			ob := &common.Account{stb, a}
			if !oa.IsEqualTo(ob) {
				ret = append(ret, a)
			}
		} else {
			ret = append(ret, a)
		}
	}

	for a, ok := range other {
		if ok {
			ret = append(ret, a)
		}
	}

	return ret
}

func CompareWithoutSuicides(sta common.State, stb common.State) []common.Address {
	diff := Compare(sta, stb)
	ret := make([]common.Address, 0, len(diff))
	for _, a := range diff {
		oa := &common.Account{sta, a}
		ob := &common.Account{stb, a}
		if !oa.HasSuicide() && !ob.HasSuicide() {
			ret = append(ret, a)
		}
	}
	return ret
}
