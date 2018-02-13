package ethash

import (
	"sync"
	"math/big"
	"time"

	"github.com/sudachen/playground/logger/glog"
	"github.com/sudachen/playground/logger"

	"github.com/ethereum/go-ethereum/pow"
	"github.com/ethereum/go-ethereum/common"
	"fmt"
)

var maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))


// Light implements the Verify half of the proof of work. It uses a few small
// in-memory caches to verify the nonces found by Full.
type Light struct {
	test bool // If set, use a smaller cache size

	mu     sync.Mutex        // Protects the per-epoch map of verification caches
	caches map[uint64]*cache // Currently maintained verification caches
	future *cache            // Pre-generated cache for the estimated future DAG

	NumCaches int // Maximum number of caches to keep before eviction (only init, don't modify)
}

func (l *Light) getCache(blockNum uint64) *cache {
	var c *cache
	epoch := blockNum / epochLength

	// If we have a PoW for that epoch, use that
	l.mu.Lock()
	if l.caches == nil {
		l.caches = make(map[uint64]*cache)
	}
	if l.NumCaches == 0 {
		l.NumCaches = 3
	}
	c = l.caches[epoch]
	if c == nil {
		// No cached DAG, evict the oldest if the cache limit was reached
		if len(l.caches) >= l.NumCaches {
			var evict *cache
			for _, cache := range l.caches {
				if evict == nil || evict.used.After(cache.used) {
					evict = cache
				}
			}
			glog.V(logger.Debug).Infof("Evicting DAG for epoch %d in favour of epoch %d", evict.epoch, epoch)
			delete(l.caches, evict.epoch)
		}
		// If we have the new DAG pre-generated, use that, otherwise create a new one
		if l.future != nil && l.future.epoch == epoch {
			glog.V(logger.Debug).Infof("Using pre-generated DAG for epoch %d", epoch)
			c, l.future = l.future, nil
		} else {
			glog.V(logger.Debug).Infof("No pre-generated DAG available, creating new for epoch %d", epoch)
			c = &cache{epoch: epoch, test: l.test}
		}
		l.caches[epoch] = c

		// If we just used up the future cache, or need a refresh, regenerate
		if l.future == nil || l.future.epoch <= epoch {
			glog.V(logger.Debug).Infof("Pre-generating DAG for epoch %d", epoch+1)
			l.future = &cache{epoch: epoch + 1, test: l.test}
			go l.future.generate()
		}
	}
	c.used = time.Now()
	l.mu.Unlock()

	// Wait for generation finish and return the cache
	c.generate()
	return c
}

// Verify checks whether the block's nonce is valid.
func (l *Light) Verify(block pow.Block) bool {
	// TODO: do ethash_quick_verify before getCache in order
	// to prevent DOS attacks.
	blockNum := block.NumberU64()
	if blockNum >= epochLength*2048 {
		glog.V(logger.Debug).Infof("block number %d too high, limit is %d", epochLength*2048)
		return false
	}

	difficulty := block.Difficulty()
	/* Cannot happen if block header diff is validated prior to PoW, but can
		 happen if PoW is checked first due to parallel PoW checking.
		 We could check the minimum valid difficulty but for SoC we avoid (duplicating)
	   Ethereum protocol consensus rules here which are not in scope of Ethash
	*/
	if difficulty.Cmp(common.Big0) == 0 {
		glog.V(logger.Debug).Infof("invalid block difficulty")
		return false
	}

	cache := l.getCache(blockNum)
	dagSize := datasize(blockNum)
	if l.test {
		dagSize = dagSizeForTesting
	}
	// Recompute the hash using the cache.
	ok, mixDigest, result := cache.compute(dagSize, block.HashNoNonce(), block.Nonce())
	if !ok {
		return false
	}

	// avoid mixdigest malleability as it's not included in a block's "hashNononce"
	if block.MixDigest() != mixDigest {
		fmt.Println("block.MixDigest() != mixDigest", block.MixDigest(), mixDigest )
		return false
	}

	// The actual check.
	target := new(big.Int).Div(maxUint256, difficulty)
	return result.Big().Cmp(target) <= 0
}

