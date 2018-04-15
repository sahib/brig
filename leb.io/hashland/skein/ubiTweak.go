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
package skein

const (
    Key             int = 0
    Config          int = 4
    Personalization int = 8
    PublicKey       int = 12
    KeyIdentifier   int = 16
    Nonce           int = 20
    Message         int = 48
    Out             int = 63
)

const t1FlagFinal = uint64(1) << 63
const t1FlagFirst = uint64(1) << 62
const t1FlagBitPad = uint64(1) << 55

type ubiTweak struct {
    tweak [2]uint64
}

func newUbiTweak() *ubiTweak {
    return new(ubiTweak)
}

/**
 * Get status of the first block flag.
 */
func (u *ubiTweak) isFirstBlock() bool {
    return (u.tweak[1] & t1FlagFirst) != 0
}

/**
 * Sets status of the first block flag.
 */
func (u *ubiTweak) setFirstBlock(value bool) {
    if value {
        u.tweak[1] |= t1FlagFirst
    } else {
        u.tweak[1] &^= t1FlagFirst
    }
}

/**
 * Gets status of the final block flag.
 */
func (u *ubiTweak) isFinalBlock() bool {
    return (u.tweak[1] & t1FlagFinal) != 0
}

/**
 * Sets status of the final block flag.
 */
func (u *ubiTweak) setFinalBlock(value bool) {
    if value {
        u.tweak[1] |= t1FlagFinal
    } else {
        u.tweak[1] &^= t1FlagFinal
    }
}

/**
 * Gets status of the final block flag.
 */
func (u *ubiTweak) isBitPad() bool {
    return (u.tweak[1] & t1FlagBitPad) != 0
}

/**
 * Sets status of the final block flag.
 */
func (u *ubiTweak) setBitPad(value bool) {
    if value {
        u.tweak[1] |= t1FlagBitPad
    } else {
        u.tweak[1] &^= t1FlagBitPad
    }
}

/**
 * Gets  the current tree level.
 */
func (u *ubiTweak) getTreeLevel() byte {
    return byte((u.tweak[1] >> 48) & 0x7f)
}

/**
 * Set the current tree level.
 * 
 * @param value
 *          the tree level
 */
func (u *ubiTweak) setTreeLevel(value int) {
    u.tweak[1] &^= uint64(0x7f) << 48
    u.tweak[1] |= uint64(value) << 48
}

/**
 * Gets the number of bytes processed so far, inclusive.
 * 
 * @return
 *      Number of processed bytes.
 */
func (u *ubiTweak) getBitsProcessed() (low, high uint64) {
    low = u.tweak[0]
    high = u.tweak[1] & 0xffffffff
    return
}

/**
 * Set the number of bytes processed so far
 * 
 * @param value
 *        The number of bits to set - low 64 bits
 */
func (u *ubiTweak) setBitsProcessed(value uint64) {
    u.tweak[0] = value
    u.tweak[1] &= 0xffffffff00000000
}

/**
 * Add number of processed bytes.
 * 
 * Adds the integer value to the 96-bit field of processed
 * bytes.
 *  
 * @param value
 *        Number of processed bytes.
 */
func (u *ubiTweak) addBytesProcessed(value int) {
    const len = 3
    carry := uint64(value)

    var words [len]uint64

    words[0] = u.tweak[0] & 0xffffffff
    words[1] = (u.tweak[0] >> 32) & 0xffffffff
    words[2] = u.tweak[1] & 0xffffffff

    for i := 0; i < len; i++ {
        carry += words[i]
        words[i] = carry
        carry >>= 32
    }
    u.tweak[0] = words[0] & 0xffffffff
    u.tweak[0] |= (words[1] & 0xffffffff) << 32
    u.tweak[1] |= words[2] & 0xffffffff
}

/**
 * Get the current UBI block type.
 */
func (u *ubiTweak) getBlockType() uint64 {
    return (u.tweak[1] >> 56) & 0x3f
}

/**
 * Set the current UBI block type.
 * 
 * @param value
 *        Block type 
 */
func (u *ubiTweak) setBlockType(value uint64) {
    u.tweak[1] = value << 56
}

/**
 * Starts a new UBI block type by setting BitsProcessed to zero, setting
 * the first flag, and setting the block type.
 *
 * @param type
 *     The UBI block type of the new block
 */
func (u *ubiTweak) startNewBlockType(t uint64) {
    u.setBitsProcessed(0)
    u.setBlockType(t)
    u.setFirstBlock(true)
}

/**
 * @return the tweak
 */
func (u *ubiTweak) getTweak() []uint64 {
    return u.tweak[:]
}

/**
 * @param word0
 *      the lower word of the tweak
 * @param word1
 *      the upper word of the tweak
 */
func (u *ubiTweak) setTweak(tw []uint64) {
    u.tweak[0] = tw[0]
    u.tweak[1] = tw[1]
}
