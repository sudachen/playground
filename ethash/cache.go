package ethash

import (
	"time"
	"sync"

	"github.com/sudachen/playground/logger/glog"
	"github.com/sudachen/playground/logger"
	"github.com/sudachen/playground/sha3"

	"github.com/ethereum/go-ethereum/common"
)

const NodeSize = 64
const NodeWords = NodeSize/4
const CacheRounds = 3
const MixBytes = 128
const MixWords = MixBytes/4
const MixNodes = MixWords/NodeWords
const EthAccess = 64
const DatasetParents = 256
const FnvPrime uint32 = 0x01000193

type node struct {
	w [NodeWords]uint32
}

type cache struct {
	epoch uint64
	used  time.Time
	test  bool

	gen sync.Once // ensures cache is only generated once.

	block uint64
	nodes []node
}

func (c *cache) new(cacheSize uint64,seed common.Hash) {
	n := make([]node,cacheSize/NodeSize)
	count := len(n)
	sha512 := new512()

	sha512.h32(&n[0],&seed)

	for i := 1; i < count; i++ {
		sha512.h64(&n[i],&n[i-1])
	}

	for j := 0; j < CacheRounds; j++ {
		for i := 0; i < count; i++ {
			idx := n[i].w[0] % uint32(count)
			d := n[(count - 1 + i) % count]
			for w := 0; w < NodeWords; w++ {
				d.w[w] ^= n[idx].w[w]
			}
			sha512.h64(&n[i],&d)
		}
	}

	c.nodes = n
}

func (c *cache) generate() {
	c.gen.Do(func() {
		started := time.Now()
		seedHash := makeSeedHash(c.epoch)
		glog.V(logger.Debug).Infof("Generating cache for epoch %d (%x)", c.epoch, seedHash)
		size := cachesize(c.epoch * epochLength)
		if c.test {
			size = cacheSizeForTesting
		}
		c.new(size,seedHash)
		glog.V(logger.Debug).Infof("Done generating cache for epoch %d, it took %v", c.epoch, time.Since(started))
	})
}

func (c *cache) dagi(index uint32, r *node, k *keccakf ) {
	n := uint32(len(c.nodes))
	*r = c.nodes[index%n]
	r.w[0] ^= index
	k.h64(r,r)
	for i := uint32(0); i < DatasetParents; i++ {
		ci := ((index ^ i) * FnvPrime ^ (r.w[i % NodeWords])) % n
		parent := &c.nodes[ci];
		dagiFNV(r, parent)
	}
	k.h64(r,r)
	return
}

func (c *cache) compute(fullSize uint64, hash common.Hash, nonce uint64) (ok bool, mixDigest, result common.Hash) {
	numMixes := uint32(fullSize/(MixWords * 4))
	var s node
	var mix [MixWords]uint32

	sha512 := new512()
	copyB32ToNode(hash[:],&s)
	s.w[8] = uint32(nonce)
	s.w[9] = uint32(nonce >> 32)
	sha512.h40(&s,&s)

	copy(mix[:NodeWords],s.w[:])
	copy(mix[NodeWords:],s.w[:])

	var r1, r2 node
	for i := uint32(0); i < EthAccess; i++ {
		p := ((s.w[0] ^ i) * FnvPrime ^ (mix[i%MixWords])) % numMixes
		c.dagi(p * MixNodes, &r1, sha512)
		c.dagi(p * MixNodes + 1, &r2, sha512)
		for w := 0; w < NodeWords; w++ {
			mix[w] = mix[w] * FnvPrime ^ r1.w[w]
			mix[w+NodeWords] = mix[w+NodeWords] * FnvPrime ^ r2.w[w]
		}
	}

	var cmix node

	for w := 0; w < MixWords; w += 4 {
		reduction := mix[w]
		reduction = reduction * FnvPrime ^ mix[w + 1]
		reduction = reduction * FnvPrime ^ mix[w + 2]
		reduction = reduction * FnvPrime ^ mix[w + 3]
		cmix.w[w/4] = reduction
	}

	copyNodeToB32(&cmix,mixDigest[:])

	q := make([]byte,64+32)
	copyNodeToB64(&s,q)
	copyNodeToB32(&cmix,q[64:])
	sha256 := sha3.NewKeccak256().(h)
	sha256.Write(q)
	copy(result[:],sha256.Out())

	ok = true

	return
}
