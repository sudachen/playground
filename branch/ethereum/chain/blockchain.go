package chain

import (
	"fmt"
	"context"
	"path/filepath"

	"github.com/sudachen/benchmark"
	"github.com/sudachen/misc/out"
	"github.com/sudachen/misc/run"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"os"
	"encoding/hex"
	"crypto/rand"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/params"
)

const DbCacheSize = 256

type Options struct {
	Identity 		string
	ChainDir      	string
	TempDbDir		string
	CacheSize		int
	ExportQueLen	int
}

func dataDir(dir, identity string) string {
	rp := filepath.Join(dir,identity)
	if !filepath.IsAbs(rp) {
		af, e := filepath.Abs(rp)
		if e != nil {
			out.Fatalf("cannot make absolute path for chain data dir: %v: %v", rp, e)
		}
		rp = af
	}
	return rp
}

func (o *Options) AbsChainDir() string {
	if o.ChainDir == "" {
		return dataDir(node.DefaultDataDir(),o.identity())
	}
	return dataDir(o.ChainDir,o.identity())
}

func (o *Options) identity() string {
	if o.Identity == "" {
		return "mainnet"
	}
	return o.Identity
}

type tdb struct {
	ethdb.Database
	path string
}

func (db *tdb) Close() {
	db.Database.Close()
	if fi, err := os.Stat(db.path); err == nil && fi.IsDir() {
		os.RemoveAll(db.path)
	}
}

func NewTempDb(o *Options) (ethdb.Database, error) {
	dbdir := o.TempDbDir
	if dbdir == "" {
		randBytes := make([]byte, 16)
		rand.Read(randBytes)
		dbdir = filepath.Join(os.TempDir(), "eth."+hex.EncodeToString(randBytes)+".chain")
	}
	if _, err := os.Stat(dbdir); err == nil {
		os.RemoveAll(dbdir)
	}

	cacheSize := o.CacheSize
	if cacheSize == 0 {
		cacheSize = DbCacheSize
	}

	db, err :=  ethdb.NewLDBDatabase(
		dbdir,
		cacheSize,
		databaseHandles())

	if err != nil {
		return nil, err
	}

	if o.TempDbDir != "" {
		return db, nil
	}
	return &tdb{db, dbdir}, nil
}

func makeChain(o *Options) (*core.BlockChain, ethdb.Database, error) {
	var err error
	cacheSize := o.CacheSize
	if cacheSize == 0 {
		cacheSize = DbCacheSize
	}
	dbpath := filepath.Join(o.AbsChainDir(),"chaindata")
	fmt.Println(dbpath)
	db, err :=  ethdb.NewLDBDatabase(
		dbpath,
		cacheSize,
		databaseHandles())

	if err != nil {
		return nil, nil, err
	}

	config, _, err := core.SetupGenesisBlock(db, nil)
	if err != nil {
		db.Close()
		return nil, nil, err
	}

	engine := ethash.NewFaker()
	vmcfg := vm.Config{}
	bc, err := core.NewBlockChain(db, config, engine, vmcfg)
	if err != nil {
		db.Close()
		return nil, nil, err
	}
	return bc, db, nil
}

func Export(
	o *Options, ctx context.Context, last uint64) (
		c chan *types.Block, 	 // blocks out channel
		cfg *params.ChainConfig, // chain config
		g *types.Block,          // genesis block
		st *state.StateDB,       // genesis state
		e error) {

	if bc, db, err := makeChain(o); err != nil {
		e = fmt.Errorf("failed to make chain: %v", err)
		return
	} else {
		if last == 0 {
			last = bc.CurrentBlock().NumberU64()
			fmt.Fprintf(os.Stderr,"last block %v\n",last)
		}

		cfg = bc.Config()
		g = bc.GetBlockByNumber(0)
		if g == nil {
			e = fmt.Errorf("failed to get genesis block")
			return
		}

		st, err = bc.StateAt(g.Root())
		if err != nil {
			e = fmt.Errorf("failed to get genesis state: %v", err)
			return
		}

		c = make(chan *types.Block,o.ExportQueLen)
		go func() {
			defer db.Close()
			defer close(c)

			for nr := uint64(1); nr <= last; nr++ {
				if run.Interrupted(ctx) {
					return
				}
				block := bc.GetBlockByNumber(nr)
				fmt.Fprintln(os.Stderr,"get block %#v\n",block)
				if block == nil {
					out.Error.Printf(
						"failed on block No %d: block not found",
						nr)
					return
				}
				c <- block
			}
		}()
	}
	return
}

func Process(
	c chan *types.Block, batchLen int, ctx context.Context,
	pf func(bs []*types.Block,ctx context.Context)error) error {

	blocks := make([]*types.Block, batchLen)

ProcessLoop:
	for {
		if run.Interrupted(ctx) {
			return run.InterruptedError
		}

		i := 0

	BatchingLoop:
		for ; i < batchLen; i++ {
			if b, ok := <- c; ok {
				blocks[i] = b
			} else {
				break BatchingLoop
			}
		}

		if i == 0 {
			break ProcessLoop
		}

		bs := blocks[:i]
		if err := pf(bs, ctx); err != nil {
			return err
		}
	}

	return nil
}

func Benchmark(
	c chan *types.Block, batchLen int,  ctx context.Context,
	t *benchmark.T,
	pf func([]*types.Block,context.Context,*benchmark.T)error) error {

	blocks := make([]*types.Block, batchLen)

	for {
		if run.Interrupted(ctx) {
			return run.InterruptedError
		}

		i := 0

		for ; i < batchLen; i++ {
			if b, ok := <- c; ok {
				blocks[i] = b
			} else {
				break
			}
		}

		bs := blocks[:i]
		if len(bs) > 0 {
			f := func(t1 *benchmark.T)error{return pf(bs,ctx,t1)}
			if err := t.Run(
				fmt.Sprintf("[%v-%v]",bs[0].NumberU64(),bs[len(bs)-1].NumberU64()),
				f); err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}
