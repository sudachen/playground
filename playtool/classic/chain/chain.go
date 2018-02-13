package chain

import (
	"path/filepath"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/pow"
	"github.com/ethereum/go-ethereum/ethdb"
)

type Options struct {
	TestNet         bool
	NetID			int
	BootNodes       []string
	PoW				pow.PoW
	DataDir         string
	CacheSize		int
	NewProcessor func (c *core.ChainConfig, bc *core.BlockChain) core.Processor
}

func (o *Options) Identity() string {
	if o.TestNet {
		return core.DefaultConfigMorden.Identity
	} else {
		return core.DefaultConfigMainnet.Identity
	}
}

func (o *Options) AbsDataDir() string {
	rp := common.EnsurePathAbsoluteOrRelativeTo(o.DataDir, o.Identity())
	if !filepath.IsAbs(rp) {
		af, e := filepath.Abs(rp)
		if e != nil {
			glog.Fatalf("cannot make absolute path for chain data dir: %v: %v", rp, e)
		}
		rp = af
	}
	return rp
}

func MakeSufficientChainConfig(o *Options) *core.SufficientChainConfig {
	config := &core.SufficientChainConfig{}
	config.Identity = o.Identity()

	if o.TestNet {
		config.Name = core.DefaultConfigMorden.Name
		config.ChainConfig = core.DefaultConfigMorden.ChainConfig
	} else {
		config.Name = core.DefaultConfigMainnet.Name
		config.ChainConfig = core.DefaultConfigMainnet.ChainConfig
	}

	if o.NetID != 0 {
		config.Network = o.NetID // 1, default mainnet
	} else {
		config.Network = eth.NetworkId
		if o.TestNet {
			config.Network += 1
		}
	}

	config.Consensus = "ethash"
	if o.TestNet {
		config.Genesis = core.DefaultConfigMorden.Genesis
	} else {
		config.Genesis = core.DefaultConfigMainnet.Genesis
	}

	if o.BootNodes != nil {
		config.ParsedBootstrap = core.ParseBootstrapNodeStrings(o.BootNodes)
	} else if o.TestNet {
		config.ParsedBootstrap = core.DefaultConfigMorden.ParsedBootstrap
	} else {
		config.ParsedBootstrap = core.DefaultConfigMainnet.ParsedBootstrap
	}

	if o.TestNet {
		state.StartingNonce = state.DefaultTestnetStartingNonce // (2**20)
	}
	return config
}

func MakeChainDatabase(o *Options) (ethdb.Database, error) {
	cacheSize := o.CacheSize
	if cacheSize == 0 {
		cacheSize = 256
	}
	return ethdb.NewLDBDatabase(filepath.Join(o.AbsDataDir(),"chaindata"), cacheSize, databaseHandles())
}

func MakeChain(o *Options, chainDb ethdb.Database) (chain *core.BlockChain) {
	var err error
	sconf := MakeSufficientChainConfig(o)

	PoW := o.PoW
	if PoW == nil {
		PoW = pow.PoW(core.FakePow{})
	}

	chain, err = core.NewBlockChain(chainDb, sconf.ChainConfig, PoW, new(event.TypeMux))
	if o.NewProcessor != nil {
		chain.SetProcessor(o.NewProcessor(sconf.ChainConfig,chain))
	}
	if err != nil {
		glog.Fatal("Could not start chainmanager: ", err)
	}
	return chain
}
