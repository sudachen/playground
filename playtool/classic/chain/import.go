package chain

import (
	"io"
	"fmt"
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/sudachen/benchmark"
	"github.com/sudachen/misc/run"
)

func BenchChainImport(o *Options, chainDb ethdb.Database, r io.Reader, importBatchSize int, ctx context.Context, t *benchmark.T) error {
	var err error
	if chainDb == nil {
		chainDb, err = MakeChainDatabase(o)
		if err != nil {
			return fmt.Errorf("Could not open database: ", err)
		}
		defer chainDb.Close()
	}
	chain := MakeChain(o,chainDb)
	s := rlp.NewStream(r, 0)
	err = BenchImportChainBatching(chain, importBatchSize, -1, s, ctx, t)
	return err
}

func DecodeChainBlocks(s *rlp.Stream, bs []*types.Block) (int, error) {
	i := 0
	for ; i < len(bs); i++ {
		var b types.Block
		if err := s.Decode(&b); err == io.EOF {
			break
		} else if err != nil {
			return i, fmt.Errorf("at block %d: %v", i, err)
		}
		// don't import first block
		if b.NumberU64() == 0 {
			i--
			continue
		}
		bs[i] = &b
	}
	return i, nil
}

const limitBatchLen = 1000

func BenchImportChainBatching(chain *core.BlockChain, importBatchSize int, batchesCount int, s *rlp.Stream, ctx context.Context, t *benchmark.T) error {
	// Watch for Ctrl-C while the import is running.
	// If a signal is received, the import will stop at the next batch.

	hasAllBlocks := func(bs []*types.Block) bool {
		for _, b := range bs {
			if !chain.HasBlock(b.Hash()) {
				return false
			}
		}
		return true
	}

	blocks := make(types.Blocks, importBatchSize)

	n := 0

	breakErr := errors.New("break")

Batching:
	for batch := 0; batchesCount < 0 || batch < batchesCount; batch++ {
		var bs types.Blocks

		err := t.Run(fmt.Sprintf("batch-%d",batch),func(t *benchmark.T) error {
			t.Start()

			// Load a batch of RLP blocks.
			if run.Interrupted(ctx) {
				return fmt.Errorf("interrupted")
			}
			if i, err := DecodeChainBlocks(s, blocks); err != nil {
				return err
			} else {
				if i == 0 {
					return breakErr // brake batching loop
				}
				n += i
				bs = blocks[:i]
			}

			// Import the batch.
			if run.Interrupted(ctx) {
				return fmt.Errorf("interrupted")
			}

			if hasAllBlocks(bs) {
				glog.D(logger.Warn).Warnf("skipping batch %d, all blocks present [%x / %x]",
					batch, bs[0].Hash().Bytes()[:4], bs[len(bs)-1].Hash().Bytes()[:4])
				return nil // continue batching loop
			}

			for len(bs) > 0 {
				if run.Interrupted(ctx) {
					return fmt.Errorf("interrupted")
				}
				var q types.Blocks
				if len(bs) > limitBatchLen {
					q = bs[:limitBatchLen]
					bs = bs[limitBatchLen:]
				} else {
					q = bs
					bs = nil
				}
				if _, err := chain.InsertChain(q); err != nil {
					return fmt.Errorf("invalid block %d: %v", n, err)
				}
			}

			return nil
		})

		switch err {
		case nil: // do nothing, continue
		case breakErr: break Batching
		default: return err
		}
	}

	return nil
}

