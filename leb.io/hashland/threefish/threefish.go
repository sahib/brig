// Copyright (C) 2011 Werner Dittmann
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// Authors: Werner Dittmann <Werner.Dittmann@t-online.de>
//

// This package implements the Threefish cipher as specified in the Skein V1.3
// specification. The Skein digest algorithm uses Threefish to generate
// the digests.
//
// NOTE: Threefish is a new cipher algorithm  - use with care until fully analysed.
//
package threefish

import (
    "strconv"
    "encoding/binary"
)

// General Threefish constants
//
const (
    KEY_SCHEDULE_CONST  = uint64(0x1BD11BDAA9FC1A22)
    EXPANDED_TWEAK_SIZE = 3
)

// Internal interface to simplify Threefish usage
//
type cipherInternal interface {
    // Encrypt function
    // 
    // Derived classes must implement this function.
    // 
    // input
    //     The plaintext input.
    // output
    //     The ciphertext output.
    //
    encrypt(input, output []uint64)

    // Decrypt function
    // 
    // Derived classes must implement this function.
    // 
    // input
    //     The ciphertext input.
    // output
    //     The plaintext output.
    //
    decrypt(input, output []uint64)

    getTempData() ([]uint64, []uint64)
    setTweak(tweak []uint64)
    setKey(key []uint64)
}

// A Cipher is an instance of Threefish using a particular key and state size.
//
type Cipher struct {
    stateSize int
    cipherInternal
}

type KeySizeError int

func (k KeySizeError) Error() string {
    return "crypto/threefish: invalid key size " + strconv.Itoa(int(k))
}

// NewCipher creates and returns a Cipher.
//
// The key length can be 32, 64 or 128 bytes and must match the Threefish
// state size. The blocksize is the same as the key length (state size).
// The tweak is a uint64 array with two elements.
//
// key
//      Key data, key length selects the internal state size
// tweak
//      The initial Tweak data for this threefish instance
//
func New(key []byte, tweak []uint64) (*Cipher, error) {
    var err error
    var internal cipherInternal

    switch len(key) {
    case 32:
        internal, err = newThreefish256(key, tweak)
    case 64:
        internal, err = newThreefish512(key, tweak)
    case 128:
        internal, err = newThreefish1024(key, tweak)
    default:
        return nil, KeySizeError(len(key))
    }
    return &Cipher{len(key) * 8, internal}, err
}

// New64 creates and returns a Cipher.
//
// The key is a uint64 array of 4, 8 or 16 elements. The key length must match the
// Threefish state size. The blocksize is the same as the key length (state size).
// The tweak is a uint64 array with two elements.
//
// key
//      Key data, key length selects the internal state size
// tweak
//      The initial Tweak data for this threefish instance
//
func New64(key, tweak []uint64) (*Cipher, error) {
    var err error
    var internal cipherInternal

    switch len(key) {
    case 4:
        internal, err = newThreefish256_64(key, tweak)
    case 8:
        internal, err = newThreefish512_64(key, tweak)
    case 16:
        internal, err = newThreefish1024_64(key, tweak)
    default:
        return nil, KeySizeError(len(key))
    }
    return &Cipher{len(key) * 8, internal}, err
}

// NewSize creates and returns a Cipher.
//
// The size argument is the requested Threefish state size
// which is also the key and block size. Supported sizes see constants section.
//
func NewSize(size int) (*Cipher, error) {
    var err error
    var internal cipherInternal

    switch size {
    case 256:
        internal, err = newThreefish256(nil, nil)
    case 512:
        internal, err = newThreefish512(nil, nil)
    case 1024:
        internal, err = newThreefish1024(nil, nil)
    default:
        return nil, KeySizeError(size)
    }
    return &Cipher{size, internal}, err
}

// BlockSize returns the cipher's block size in bytes.
//
func (c *Cipher) BlockSize() int {
    return c.stateSize / 8
}

// Encrypt a block.
// Dst and src may point at the same memory.
//
// dst
//      Destination of encypted data (cipher data)
// src
//      Contains a block of plain data
//
func (c *Cipher) Encrypt(dst, src []byte) {

    uintLen := c.stateSize / 64

    // This saves makes
    tmpin, tmpout := c.getTempData()

    for i := 0; i < uintLen; i++ {
        tmpin[i] = binary.LittleEndian.Uint64(src[i*8 : i*8+8])
    }
    c.encrypt(tmpin, tmpout)

    for i := 0; i < uintLen; i++ {
        binary.LittleEndian.PutUint64(dst[i*8:i*8+8], tmpout[i])
    }
}

// Decrypt a block.
// Dst and src may point at the same memory.
//
// dst
//      Destination of encypted data (cipher data)
// src
//      Contains a block of plain data
//
func (c *Cipher) Decrypt(dst, src []byte) {

    uintLen := c.stateSize / 64

    // This saves a make because tmpin and tmpout are of variable length
    tmpin, tmpout := c.getTempData()

    for i := 0; i < uintLen; i++ {
        tmpin[i] = binary.LittleEndian.Uint64(src[i*8 : i*8+8])
    }
    c.decrypt(tmpin, tmpout)

    for i := 0; i < uintLen; i++ {
        binary.LittleEndian.PutUint64(dst[i*8:i*8+8], tmpout[i])
    }
}

// Encrypt a block.
// Blocks are unit64 arrays.
// Dst and src may point at the same memory.
//
// dst
//      Destination of encypted data (cipher data)
// src
//      Contains a block of plain data
//
func (c *Cipher) Encrypt64(dst, src []uint64) {
    c.encrypt(src, dst)
}

// Decrypt a block.
// Blocks are unit64 arrays.
// Dst and src may point at the same memory.
//
// dst
//      Destination of decrypted data (plain data)
// src
//      Contains a block of encrypted data (cipher data)
//
func (c *Cipher) Decrypt64(dst, src []uint64) {
    c.decrypt(src, dst)
}

// Set the tweak data.
//
// The tweak is a uint64 array with two elements.
//
func (c *Cipher) SetTweak(tweak []uint64) {
    c.setTweak(tweak)
}

// Set the key.
//
// The key must have the same length as the Threefish state size.
// 
func (c *Cipher) SetKey(key []uint64) {
    c.setKey(key)
}

// Some helper functions available for all Threefish* implementations
func setTweak(tweak, expandedTweak []uint64) {
    if tweak != nil {
        expandedTweak[0] = tweak[0]
        expandedTweak[1] = tweak[1]
        expandedTweak[2] = tweak[0] ^ tweak[1]
    }
}

func setKey(key, expandedKey []uint64) {
    var i int
    parity := uint64(KEY_SCHEDULE_CONST)

    for i = 0; i < len(expandedKey)-1; i++ {
        expandedKey[i] = key[i]
        parity ^= key[i]
    }
    expandedKey[i] = parity
}
