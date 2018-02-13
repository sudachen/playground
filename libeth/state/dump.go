package state

import (
	"fmt"
	"io"
	"github.com/sudachen/playground/libeth"
	"github.com/ethereum/go-ethereum/common"
)

func WriteDump(wr io.Writer, st libeth.State, pfx string) {
	addresses := st.Addresses(false)
	fmt.Fprintf(wr, "%sSTATE HAS %d ACCOUNTS:\n", pfx, len(addresses))
	for i, a := range addresses {
		o := libeth.Account{st, a}
		if i != 0 {
			fmt.Fprintf(wr,"%s--------\n",pfx)
		}
		fmt.Fprintf(wr, "%s%d\tAddress: %v\n", pfx, i, o.Address.Hex())
		fmt.Fprintf(wr, "%s\tBalance: %v\n", pfx, o.Balance())
		fmt.Fprintf(wr, "%s\tNonce: %v\n", pfx, o.Nonce())
		fmt.Fprintf(wr, "%s\tSuicided: %v\n", pfx, o.HasSuicide())
		fmt.Fprintf(wr, "%s\tCode: %v\n", pfx, common.ToHex(o.Code()))
		fmt.Fprintf(wr, "%s\tHash: %v\n", pfx, o.CodeHash().Hex())
		o.ProcessValues(func(k, v common.Hash) error {
			fmt.Fprintf(wr, "%s\t\t%v => %v\n", pfx, k.Hex(), v.Hex())
			return nil
		}, false)
	}
}

func WriteDiff(wr io.Writer, st libeth.State, etalonSt libeth.State, addresses ...common.Address) {
	for _, a := range addresses {
		o := libeth.Account{st, a}
		e := libeth.Account{etalonSt, a}
		if o.Exists() != e.Exists() {
			if !o.HasSuicide() && !e.HasSuicide() {
				var oExists= "exists"
				var eExists= "it have not"
				if !o.Exists() {
					oExists = "does not exist"
					eExists = "it have to be"
				}
				fmt.Fprintf(wr,"address %s %s but %s\n",
					a.Hex(), oExists, eExists)
			}
		} else {
			if o.Balance().Cmp(e.Balance()) != 0 {
				fmt.Fprintf(wr,"balance of %s is %v but have to be %v\n",
					a.Hex(), o.Balance(), e.Balance())
			}
			if o.Nonce() != e.Nonce() {
				fmt.Fprintf(wr, "nonce of %s is %v but have to be %v\n",
					a.Hex(), o.Nonce(), e.Nonce())
			}
		}
	}
}

func Compare(sta libeth.State, stb libeth.State) []common.Address {
	staa := sta.Addresses(false)
	ret := make([]common.Address, 0, len(staa))

	other := make(map[common.Address]bool)
	for _, a := range staa {
		other[a] = true
	}

	for _, a := range stb.Addresses(false) {
		if _, exists := other[a]; exists {
			other[a] = false
			oa := &libeth.Account{sta, a}
			ob := &libeth.Account{stb, a}
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

func CompareWithoutSuicides(sta libeth.State, stb libeth.State) []common.Address {
	diff := Compare(sta, stb)
	ret := make([]libeth.Address, 0, len(diff))
	for _, a := range diff {
		oa := &libeth.Account{sta, a}
		ob := &libeth.Account{stb, a}
		if !oa.HasSuicide() && !ob.HasSuicide() {
			ret = append(ret, a)
		}
	}
	return ret
}
