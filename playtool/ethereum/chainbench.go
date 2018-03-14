package ethereum

import (
	"math/big"
	"context"

	"github.com/sudachen/playground/branch/ethereum/chain"
	"github.com/sudachen/playground/libeth"
	"github.com/sudachen/benchmark"
	"github.com/sudachen/misc/run"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/common"
)

func ChainBench(o *chain.Options, batchLen int, last uint64, t *benchmark.T, newVM func()libeth.VM1) error {
	return run.WithCancelByInterruptErr(func(ctx context.Context)error{
		c, cfg, _, st, err := chain.Export(o,ctx,last)
		if err != nil {
			return err
		}

		db, err := chain.NewTempDb(o)
		if err != nil {
			return err
		}

		cache := state.NewDatabase(db)

		root, err := st.Copy().CommitTo(db,false)
		if err != nil {
			return err
		}

		hashfn := func (n uint64) common.Hash {
			b, err := db.Get(common.BigToHash(new(big.Int).SetUint64(n)).Bytes())
			if err != nil || b == nil {
				return common.Hash{}
			}
			return common.BytesToHash(b)
		}

		return chain.Benchmark(c, batchLen, ctx, t,
			func (bs []*types.Block, ctx context.Context, t1 *benchmark.T) error{
				t1.Start()
				for _, block := range bs {
					sdb, err := state.New(root,cache)
					if err != nil {
						return err
					}
					h := block.Header()
					bi := &libeth.BlockInfo{
						Header: *h,
						Blockhash: hashfn,
						Config: cfg,
					}
					for _, t := range block.Transactions() {
						m, err := t.AsMessage(types.MakeSigner(bi.Config, bi.Header.Number))
						if err != nil {
							return err
						}
						newVM().Execute(m, bi, sdb)
					}
					root, err := sdb.CommitTo(db,true)
					if root != block.Root() {
						panic("result state is not mached to the same in the blockchain")
					}
					db.Put(common.BigToHash(block.Number()).Bytes(),block.Header().Hash().Bytes())
				}
				return nil
			})
	})
}

