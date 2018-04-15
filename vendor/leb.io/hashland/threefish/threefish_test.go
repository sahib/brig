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
package threefish

import (
    "encoding/binary"
    "encoding/hex"
    "fmt"
    "testing"
    "bytes"
)

// The zeroized test data and the expected result
var three_256_00_key = []uint64{0, 0, 0, 0}

var three_256_00_input = []uint64{0, 0, 0, 0}

var three_256_00_tweak = []uint64{0, 0}

var three_256_00_result = []uint64{0x94EEEA8B1F2ADA84, 0xADF103313EAE6670,
    0x952419A1F4B16D53, 0xD83F13E63C9F6B11}

var three_256_01_key = []uint64{0x1716151413121110, 0x1F1E1D1C1B1A1918,
    0x2726252423222120, 0x2F2E2D2C2B2A2928}

var three_256_01_input = []uint64{0xF8F9FAFBFCFDFEFF, 0xF0F1F2F3F4F5F6F7,
    0xE8E9EAEBECEDEEEF, 0xE0E1E2E3E4E5E6E7}

var three_256_01_tweak = []uint64{0x0706050403020100, 0x0F0E0D0C0B0A0908}

var three_256_01_result = []uint64{0x277610F5036C2E1F, 0x25FB2ADD1267773E,
    0x9E1D67B3E4B06872, 0x3F76BC7651B39682}

var three_512_00_key = []uint64{0, 0, 0, 0, 0, 0, 0, 0}

var three_512_00_input = []uint64{0, 0, 0, 0, 0, 0, 0, 0}

var three_512_00_tweak = []uint64{0, 0}

var three_512_00_result = []uint64{0xBC2560EFC6BBA2B1, 0xE3361F162238EB40,
    0xFB8631EE0ABBD175, 0x7B9479D4C5479ED1, 0xCFF0356E58F8C27B,
    0xB1B7B08430F0E7F7, 0xE9A380A56139ABF1, 0xBE7B6D4AA11EB47E}

var three_512_01_key = []uint64{0x1716151413121110, 0x1F1E1D1C1B1A1918,
    0x2726252423222120, 0x2F2E2D2C2B2A2928, 0x3736353433323130,
    0x3F3E3D3C3B3A3938, 0x4746454443424140, 0x4F4E4D4C4B4A4948}

var three_512_01_input = []uint64{0xF8F9FAFBFCFDFEFF, 0xF0F1F2F3F4F5F6F7,
    0xE8E9EAEBECEDEEEF, 0xE0E1E2E3E4E5E6E7, 0xD8D9DADBDCDDDEDF,
    0xD0D1D2D3D4D5D6D7, 0xC8C9CACBCCCDCECF, 0xC0C1C2C3C4C5C6C7}

var three_512_01_tweak = []uint64{0x0706050403020100, 0x0F0E0D0C0B0A0908}

var three_512_01_result = []uint64{0xD4A32EDD6ABEFA1C, 0x6AD5C4252C3FF743,
    0x35AC875BE2DED68C, 0x99A6C774EA5CD06C, 0xDCEC9C4251D7F4F8,
    0xF5761BCB3EF592AF, 0xFCABCB6A3212DF60, 0xFD6EDE9FF9A2E14E}

var three_1024_00_key = []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

var three_1024_00_input = []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

var three_1024_00_tweak = []uint64{0, 0}

var three_1024_00_result = []uint64{0x04B3053D0A3D5CF0, 0x0136E0D1C7DD85F7,
    0x067B212F6EA78A5C, 0x0DA9C10B4C54E1C6, 0x0F4EC27394CBACF0,
    0x32437F0568EA4FD5, 0xCFF56D1D7654B49C, 0xA2D5FB14369B2E7B,
    0x540306B460472E0B, 0x71C18254BCEA820D, 0xC36B4068BEAF32C8,
    0xFA4329597A360095, 0xC4A36C28434A5B9A, 0xD54331444B1046CF,
    0xDF11834830B2A460, 0x1E39E8DFE1F7EE4F}

var three_1024_01_key = []uint64{0x1716151413121110, 0x1F1E1D1C1B1A1918,
    0x2726252423222120, 0x2F2E2D2C2B2A2928, 0x3736353433323130,
    0x3F3E3D3C3B3A3938, 0x4746454443424140, 0x4F4E4D4C4B4A4948,
    0x5756555453525150, 0x5F5E5D5C5B5A5958, 0x6766656463626160,
    0x6F6E6D6C6B6A6968, 0x7776757473727170, 0x7F7E7D7C7B7A7978,
    0x8786858483828180, 0x8F8E8D8C8B8A8988}

var three_1024_01_input = []uint64{0xF8F9FAFBFCFDFEFF, 0xF0F1F2F3F4F5F6F7,
    0xE8E9EAEBECEDEEEF, 0xE0E1E2E3E4E5E6E7, 0xD8D9DADBDCDDDEDF,
    0xD0D1D2D3D4D5D6D7, 0xC8C9CACBCCCDCECF, 0xC0C1C2C3C4C5C6C7,
    0xB8B9BABBBCBDBEBF, 0xB0B1B2B3B4B5B6B7, 0xA8A9AAABACADAEAF,
    0xA0A1A2A3A4A5A6A7, 0x98999A9B9C9D9E9F, 0x9091929394959697,
    0x88898A8B8C8D8E8F, 0x8081828384858687}

var three_1024_01_tweak = []uint64{0x0706050403020100, 0x0F0E0D0C0B0A0908}

var three_1024_01_result = []uint64{0x483AC62C27B09B59, 0x4CB85AA9E48221AA,
    0x80BC1644069F7D0B, 0xFCB26748FF92B235, 0xE83D70243B5D294B,
    0x316A3CA3587A0E02, 0x5461FD7C8EF6C1B9, 0x7DD5C1A4C98CA574,
    0xFDA694875AA31A35, 0x03D1319C26C2624C, 0xA2066D0DF2BF7827,
    0x6831CCDAA5C8A370, 0x2B8FCD9189698DAC, 0xE47818BBFD604399,
    0xDF47E519CBCEA541, 0x5EFD5FF4A5D4C259}

var key, dataIn, dataOut, result []byte

func TestThreefish(t *testing.T) {
    basicTest256()
    basicTest512()
    basicTest1024()

}

func basicTest256() bool {
    stateSize := 256

    // Key must match Threefish state size
    key = make([]byte, stateSize/8)

    // For simple ECB mode length matches state size
    dataIn = make([]byte, stateSize/8)
    dataOut = make([]byte, stateSize/8)
    result = make([]byte, stateSize/8)

    // Prepare first test vector as byte array
    for i := 0; i < len(three_256_00_input); i++ {
        binary.LittleEndian.PutUint64(dataIn[i*8:i*8+8], three_256_00_input[i])
        binary.LittleEndian.PutUint64(key[i*8:i*8+8], three_256_00_key[i])
        binary.LittleEndian.PutUint64(result[i*8:i*8+8], three_256_00_result[i])
    }
    // Create cipher with key and tweak data
    cipher, _ := New(key[:], three_256_00_tweak[:])

    // Encrypt and check
    cipher.Encrypt(dataOut[:], dataIn[:])
    if ret := bytes.Compare(dataOut[:], result[:]); ret != 0 {
        fmt.Printf("Wrong cipher text 256 00:\n%s\n", hex.EncodeToString(dataOut))
        return false
    }
    // Decrypt and check
    cipher.Decrypt(result[:], dataOut[:])
    if ret := bytes.Compare(dataIn[:], result[:]); ret != 0 {
        fmt.Printf("Decrypt failed 256 00:\n%s\n", hex.EncodeToString(result))
        return false
    }
    // Prepare next test vector as byte array
    for i := 0; i < len(three_256_00_input); i++ {
        binary.LittleEndian.PutUint64(dataIn[i*8:i*8+8], three_256_01_input[i])
        binary.LittleEndian.PutUint64(key[i*8:i*8+8], three_256_01_key[i])
        binary.LittleEndian.PutUint64(result[i*8:i*8+8], three_256_01_result[i])
    }
    // Create cipher with key and tweak data
    cipher, _ = New(key[:], three_256_01_tweak[:])

    // Encrypt and check
    cipher.Encrypt(dataOut[:], dataIn[:])

    // plaintext feed forward
    for i := 0; i < len(dataIn); i++ {
        dataOut[i] ^= dataIn[i]
    }
    if ret := bytes.Compare(dataOut[:], result[:]); ret != 0 {
        fmt.Printf("Wrong cipher text 256 01:\n%s\n", hex.EncodeToString(dataOut))
        return false
    }
    // Decrypt and check
    // plaintext feed backward :-)
    for i := 0; i < len(dataIn); i++ {
        dataOut[i] ^= dataIn[i]
    }
    cipher.Decrypt(result[:], dataOut[:])
    if ret := bytes.Compare(dataIn[:], result[:]); ret != 0 {
        fmt.Printf("Decrypt failed 256 01:\n%s\n", hex.EncodeToString(result))
        return false
    }

    return true
}

func basicTest512() bool {
    stateSize := 512

    // Key must match Threefish state size
    key = make([]byte, stateSize/8)

    // For simple ECB mode length matches state size
    dataIn = make([]byte, stateSize/8)
    dataOut = make([]byte, stateSize/8)
    result = make([]byte, stateSize/8)

    // Prepare first test vector as byte array
    for i := 0; i < len(three_512_00_input); i++ {
        binary.LittleEndian.PutUint64(dataIn[i*8:i*8+8], three_512_00_input[i])
        binary.LittleEndian.PutUint64(key[i*8:i*8+8], three_512_00_key[i])
        binary.LittleEndian.PutUint64(result[i*8:i*8+8], three_512_00_result[i])
    }
    // Create cipher with key and tweak data
    cipher, _ := New(key[:], three_512_00_tweak[:])

    // Encrypt and check
    cipher.Encrypt(dataOut[:], dataIn[:])
    if ret := bytes.Compare(dataOut[:], result[:]); ret != 0 {
        fmt.Printf("Wrong cipher text 512 00:\n%s\n", hex.EncodeToString(dataOut))
        return false
    }
    // Decrypt and check
    cipher.Decrypt(result[:], dataOut[:])
    if ret := bytes.Compare(dataIn[:], result[:]); ret != 0 {
        fmt.Printf("Decrypt failed 512 00:\n%s\n", hex.EncodeToString(result))
        return false
    }
    // Prepare next test vector as byte array
    for i := 0; i < len(three_512_00_input); i++ {
        binary.LittleEndian.PutUint64(dataIn[i*8:i*8+8], three_512_01_input[i])
        binary.LittleEndian.PutUint64(key[i*8:i*8+8], three_512_01_key[i])
        binary.LittleEndian.PutUint64(result[i*8:i*8+8], three_512_01_result[i])
    }
    // Create cipher with key and tweak data
    cipher, _ = New(key[:], three_512_01_tweak[:])

    // Encrypt and check
    cipher.Encrypt(dataOut[:], dataIn[:])

    // plaintext feed forward
    for i := 0; i < len(dataIn); i++ {
        dataOut[i] ^= dataIn[i]
    }
    if ret := bytes.Compare(dataOut[:], result[:]); ret != 0 {
        fmt.Printf("Wrong cipher text 512 01:\n%s\n", hex.EncodeToString(dataOut))
        return false
    }
    // Decrypt and check
    // plaintext feed backward :-)
    for i := 0; i < len(dataIn); i++ {
        dataOut[i] ^= dataIn[i]
    }
    cipher.Decrypt(result[:], dataOut[:])
    if ret := bytes.Compare(dataIn[:], result[:]); ret != 0 {
        fmt.Printf("Decrypt failed 512 01:\n%s\n", hex.EncodeToString(result))
        return false
    }

    return true
}

func basicTest1024() bool {
    stateSize := 1024

    // Key must match Threefish state size
    key = make([]byte, stateSize/8)

    // For simple ECB mode length matches state size
    dataIn = make([]byte, stateSize/8)
    dataOut = make([]byte, stateSize/8)
    result = make([]byte, stateSize/8)

    // Prepare first test vector as byte array
    for i := 0; i < len(three_1024_00_input); i++ {
        binary.LittleEndian.PutUint64(dataIn[i*8:i*8+8], three_1024_00_input[i])
        binary.LittleEndian.PutUint64(key[i*8:i*8+8], three_1024_00_key[i])
        binary.LittleEndian.PutUint64(result[i*8:i*8+8], three_1024_00_result[i])
    }
    // Create cipher with key and tweak data
    cipher, _ := New(key[:], three_1024_00_tweak[:])

    // Encrypt and check
    cipher.Encrypt(dataOut[:], dataIn[:])
    if ret := bytes.Compare(dataOut[:], result[:]); ret != 0 {
        fmt.Printf("Wrong cipher text 1024 00:\n%s\n", hex.EncodeToString(dataOut))
        return false
    }
    // Decrypt and check
    cipher.Decrypt(result[:], dataOut[:])
    if ret := bytes.Compare(dataIn[:], result[:]); ret != 0 {
        fmt.Printf("Decrypt failed 1024 00:\n%s\n", hex.EncodeToString(result))
        return false
    }
    // Prepare next test vector as byte array
    for i := 0; i < len(three_1024_00_input); i++ {
        binary.LittleEndian.PutUint64(dataIn[i*8:i*8+8], three_1024_01_input[i])
        binary.LittleEndian.PutUint64(key[i*8:i*8+8], three_1024_01_key[i])
        binary.LittleEndian.PutUint64(result[i*8:i*8+8], three_1024_01_result[i])
    }
    // Create cipher with key and tweak data
    cipher, _ = New(key[:], three_1024_01_tweak[:])

    // Encrypt and check
    cipher.Encrypt(dataOut[:], dataIn[:])

    // plaintext feed forward
    for i := 0; i < len(dataIn); i++ {
        dataOut[i] ^= dataIn[i]
    }
    if ret := bytes.Compare(dataOut[:], result[:]); ret != 0 {
        fmt.Printf("Wrong cipher text 1024 01:\n%s\n", hex.EncodeToString(dataOut))
        return false
    }
    // Decrypt and check
    // plaintext feed backward :-)
    for i := 0; i < len(dataIn); i++ {
        dataOut[i] ^= dataIn[i]
    }
    cipher.Decrypt(result[:], dataOut[:])
    if ret := bytes.Compare(dataIn[:], result[:]); ret != 0 {
        fmt.Printf("Decrypt failed 1024 01:\n%s\n", hex.EncodeToString(result))
        return false
    }

    return true
}
