package ethash

import (
	"hash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sudachen/playground/sha3"
	"unsafe"
)

type h interface {
	hash.Hash
	Out() []byte
}

type keccakf struct {
	h
	b []byte
}

func copyNodeToB64(src *node, dst []byte) {
	*(*node)(unsafe.Pointer(&dst[0])) = *src
}

type c32 [32]byte
func copyNodeToB32(src *node, dst []byte) {
	copy(dst,(*c32)(unsafe.Pointer(src))[:])
}

func copyB64ToNode(src []byte, dst *node) {
	*dst = *(*node)(unsafe.Pointer(&src[0]))
}

type b40 [10]uint32
func copyB40ToNode(src []byte, dst *node) {
	s := (*b40)(unsafe.Pointer(&src[0]))
	d := (*b40)(unsafe.Pointer(&dst.w[0]))
	*d = *s
}

type b32 [8]uint32
func copyB32ToNode(src []byte, dst *node) {
	s := (*b32)(unsafe.Pointer(&src[0]))
	d := (*b32)(unsafe.Pointer(&dst.w[0]))
	*d = *s
}

func (h *keccakf) h64(out *node, in *node) {
	h.Reset()
	copyNodeToB64(in,h.b)
	h.Write(h.b)
	copyB64ToNode(h.Out(),out)
}

func (h *keccakf) h32(out *node, in *common.Hash) {
	h.Reset()
	h.Write(in[:])
	copyB64ToNode(h.Out(),out)
}

func (h *keccakf) h40(out *node, in *node) {
	h.Reset()
	copyNodeToB64(in,h.b)
	h.Write(h.b[0:40])
	copyB64ToNode(h.Out(),out)
}

func new512() *keccakf {
	return &keccakf{h:sha3.NewKeccak512().(h),b:make([]byte,64)}
}
