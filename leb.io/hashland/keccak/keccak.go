package keccak

// #include "KeccakNISTInterface.h"
import "C"
import (
    "hash"
    "errors"
    "unsafe"
)

type keccak struct {
    hashState  C.hashState
    bitlen     C.int
    dataLength C.DataLength
}

func newKeccak(bitlen int) hash.Hash {
    k := &keccak{bitlen: C.int(bitlen)}
    k.Reset()
    return k
}

/*
func NewCustom(bits, rounds int) hash.Hash {
    return newKeccak(bits, rounds)
}
*/

func New224() hash.Hash {
    return newKeccak(224)
}

func New256() hash.Hash {
    return newKeccak(256)
}

func New384() hash.Hash {
    return newKeccak(384)
}

func New512() hash.Hash {
    return newKeccak(512)
}

func (k *keccak) Write(b []byte) (int, error) {
    n := len(b)
    if n == 0 {
        return 0, nil
    }
    dl := C.DataLength(n*8)
    p := (*C.BitSequence)(unsafe.Pointer(&b[0]))
    if C.Update(&k.hashState, p, dl) != C.SUCCESS {
        return 0, errors.New("keccak write error")
    }
    return n, nil
}

func (k *keccak) Sum(b []byte) []byte {
    k0 := *k
    buf := make([]byte, k.Size(), k.Size())
    p := (*C.BitSequence)(unsafe.Pointer(&buf[0]))
    if C.Final(&k0.hashState, p) != C.SUCCESS {
        panic("keccak sum error")
    }
    return append(b, buf...)
}

func (k *keccak) Reset() {
    C.Init(&k.hashState, k.bitlen)
}

func (k keccak) BlockSize() int {
    return 200 - 2*k.Size()
}

func (k keccak) Size() int {
    return int(k.bitlen) / 8
}
