package channel

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Flag struct {
	// aData holds the bitset, most significant bits at lower indices.
	aData []uint32
}

// Version/collection bit lengths
const (
	Ver_1  = 32
	Ver_14 = 64

	Common_MonsterCollection = 192 // MonsterCollectionInfo
	Common_ItemCollection    = 320 // ItemCategoryInfo
	Common_PetTemplate       = 128
	Common_MobStat           = 128

	Unknown = 65536

	Default = Ver_14 // Most common versions used are UINT128.
)

// NewFlag returns a flag with Default bits (128).
func NewFlag() *Flag {
	return NewFlagBits(Default)
}

// NewFlagBits constructs a Flag with exactly uBits capacity.
func NewFlagBits(uBits int) *Flag {
	if uBits < 0 {
		uBits = 0
	}
	words := uBits >> 5 // divide by 32
	f := &Flag{aData: make([]uint32, words)}
	f.SetValue(0)
	return f
}

func NewFlagCopy(uValue *Flag, uNumBits int) *Flag {
	if uValue == nil {
		return NewFlagBits(0)
	}
	totalBits := 32 * len(uValue.aData)
	if uNumBits < 0 {
		uNumBits = 0
	}
	if uNumBits > totalBits {
		uNumBits = totalBits
	}

	out := NewFlagBits(32 * len(uValue.aData))

	// Copy the 32-bit chunks.
	fullWords := uNumBits >> 5 // floor(uNumBits / 32)
	for i := fullWords; i > 0; i-- {
		out.aData[i-1] = uValue.aData[i-1]
	}

	// Copy the remaining bits one by one.
	for i := 32 * fullWords; i < uNumBits; i++ {
		out.SetBitNumber(i, uValue.GetBitNumber(i))
	}

	// Pad remaining bits with pseudo-random values as per Java LCG-ish approach.
	if uNumBits < totalBits {
		// Java used: ((214013 * rand(0..32766) + 2531011) >> 16) & 0x7FFF
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := uNumBits; i < totalBits; i++ {
			uRand := ((214013*uint32(r.Intn(32767)) + 2531011) >> 16) & 0x7FFF
			out.SetBitNumber(i, int(uRand%2))
		}
	}

	return out
}

// CompareTo compares two flags lexicographically by their 32-bit words.
// Returns -1 if this < other, 1 if this > other, 0 if equal.
// Note: Requires equal lengths to be meaningful; differing lengths will
// compare up to min length and then treat missing words as 0.
func (f *Flag) CompareTo(other *Flag) int {
	if f == nil && other == nil {
		return 0
	}
	if f == nil {
		return -1
	}
	if other == nil {
		return 1
	}
	// Compare up to the shorter length first.
	n := len(f.aData)
	if len(other.aData) < n {
		n = len(other.aData)
	}
	for i := 0; i < n; i++ {
		if f.aData[i] < other.aData[i] {
			return -1
		}
		if f.aData[i] > other.aData[i] {
			return 1
		}
	}
	// If lengths differ, treat extra words as significance.
	if len(f.aData) < len(other.aData) {
		// If any remaining word in other is non-zero, f < other.
		for i := n; i < len(other.aData); i++ {
			if other.aData[i] != 0 {
				return -1
			}
		}
	} else if len(f.aData) > len(other.aData) {
		for i := n; i < len(f.aData); i++ {
			if f.aData[i] != 0 {
				return 1
			}
		}
	}
	return 0
}

// CompareToInt compares the entire flag to a single 32-bit integer value,
// assuming that only the least significant 32 bits (last word) may be set.
// Returns -1 if this < value, 1 if this > value, 0 if equal.
func (f *Flag) CompareToInt(value uint32) int {
	if len(f.aData) == 0 {
		if value == 0 {
			return 0
		}
		return -1
	}
	uLen := len(f.aData) - 1
	if f.aData[uLen] > value {
		return 1
	}
	for i := 0; i < uLen; i++ {
		if f.aData[i] != 0 {
			return 1
		}
	}
	if f.aData[uLen] < value {
		return -1
	}
	return 0
}

// DecodeBuffer fills the flag from a provided 32-bit decoder function.
// The decoder should return the next 4-byte unsigned value in the
// same order as the original wire format.
func (f *Flag) DecodeBuffer(decode4 func() uint32) {
	if f == nil || decode4 == nil {
		return
	}
	for i := 0; i < len(f.aData); i++ {
		f.aData[i] = decode4()
	}
}

// GetBitNumber gets the bit at position uBit (0-based), returning 0 or 1.
// Bits are addressed MSB-first per 32-bit word.
func (f *Flag) GetBitNumber(uBit int) int {
	if f == nil {
		return 0
	}
	totalBits := 32 * len(f.aData)
	if uBit < 0 || uBit >= totalBits {
		return 0
	}
	word := uBit >> 5               // uBit / 32
	off := uint(31 - (uBit & 0x1F)) // MSB-first
	return int((f.aData[word] >> off) & 1)
}

// Data returns the internal data slice (do not modify).
func (f *Flag) Data() []uint32 {
	if f == nil {
		return nil
	}
	return f.aData
}

// IsSet returns true if any bit is non-zero.
func (f *Flag) IsSet() bool {
	if f == nil {
		return false
	}
	i := 0
	uLen := len(f.aData) - 1
	if uLen < 0 {
		return false
	}
	for f.aData[i] == 0 {
		i++
		if i >= uLen {
			return f.aData[uLen] != 0
		}
	}
	return true
}

// IsZero returns true if all bits are zero.
func (f *Flag) IsZero() bool {
	if f == nil {
		return true
	}
	i := len(f.aData) - 1
	if i < 0 {
		return true
	}
	for f.aData[i] == 0 {
		i--
		if i < 0 {
			return true
		}
	}
	return false
}

// OperatorAND computes a bitwise AND and returns a new Flag with the result.
// If lengths differ, it returns nil (mirrors Java behavior).
func (f *Flag) OperatorAND(other *Flag) *Flag {
	if f == nil || other == nil {
		return nil
	}
	if len(f.aData) != len(other.aData) {
		return nil
	}
	n := len(f.aData)
	out := NewFlagBits(32 * n)
	for i := n - 1; i >= 0; i-- {
		out.aData[i] = f.aData[i] & other.aData[i]
	}
	return out
}

// IsEqual returns true if the flag equals the given scalar 32-bit value,
// assuming only the last word can carry the value.
func (f *Flag) IsEqual(value uint32) bool {
	if f == nil {
		return value == 0
	}
	i := 0
	uLen := len(f.aData) - 1
	if uLen < 0 {
		return value == 0
	}
	for f.aData[i] == 0 {
		i++
		if i >= uLen {
			uData := f.aData[uLen]
			if uData <= value {
				return uData >= value
			}
			return false
		}
	}
	return false
}

// PerformOR applies an in-place OR with another flag.
// If lengths differ, it does nothing (mirrors Java behavior).
func (f *Flag) PerformOR(other *Flag) {
	if f == nil || other == nil {
		return
	}
	if len(f.aData) != len(other.aData) {
		return
	}
	// Early-out if 'other' is all zeros.
	i := 0
	uLen := len(other.aData) - 1
	for other.aData[i] == 0 {
		i++
		if i >= uLen {
			if other.aData[uLen] == 0 {
				return
			}
			break
		}
	}
	for j := uLen; j >= 0; j-- {
		f.aData[j] |= other.aData[j]
	}
}

// OperatorOR returns a new Flag which is the bitwise OR of two flags.
// Returns nil if lengths differ (mirrors Java behavior).
func (f *Flag) OperatorOR(other *Flag) *Flag {
	if f == nil || other == nil {
		return nil
	}
	if len(f.aData) != len(other.aData) {
		return nil
	}
	uLen := len(f.aData)
	out := NewFlagBits(32 * uLen)
	for i := 0; i < uLen; i++ {
		out.aData[i] = f.aData[i] | other.aData[i]
	}
	return out
}

// ShiftLeft shifts the entire flag left by uBits (big number left shift).
func (f *Flag) ShiftLeft(uBits int) {
	if f == nil || uBits == 0 || f.IsZero() {
		return
	}
	uLen := len(f.aData)
	totalBits := 32 * uLen
	if uBits >= totalBits {
		f.SetValue(0)
		return
	}
	wordShift := uBits >> 5      // uBits / 32
	bitShift := uint(uBits & 31) // remaining bit shift

	dst := make([]uint32, uLen)
	for i := uLen - 1; i >= 0; i-- {
		src := i - wordShift
		if src < 0 {
			dst[i] = 0
			continue
		}
		val := f.aData[src] << bitShift
		if bitShift != 0 && (src-1) >= 0 {
			val |= f.aData[src-1] >> (32 - bitShift)
		}
		dst[i] = val
	}
	copy(f.aData, dst)
}

// SetBitNumber sets a specific bit to 0 or 1.
// Bits are addressed MSB-first per 32-bit word.
func (f *Flag) SetBitNumber(uBit int, uValue int) {
	if f == nil {
		return
	}
	totalBits := 32 * len(f.aData)
	if uBit < 0 || uBit >= totalBits {
		return
	}
	word := uBit >> 5
	mask := uint32(1) << uint(31-(uBit&0x1F))
	f.aData[word] |= mask
	if uValue == 0 {
		f.aData[word] = f.aData[word] ^ mask
	}
}

// SetData is a convenience that sets a single '1' bit at position uBits.
func (f *Flag) SetData(uBits int) {
	if f == nil || len(f.aData) == 0 {
		return
	}
	uLen := len(f.aData) - 1
	nIndex := uLen - (uBits >> 5)
	if nIndex < 0 || nIndex >= len(f.aData) {
		return
	}
	nValue := uint32(1) << uint(0x1F-(uBits&0x1F))
	f.aData[nIndex] |= nValue
}

// SetValue assigns a scalar 32-bit value to the least significant word,
// zeroing all higher words (mirrors Java).
func (f *Flag) SetValue(uValue uint32) {
	if f == nil {
		return
	}
	if len(f.aData) == 0 {
		return
	}
	uLen := len(f.aData) - 1
	for i := 0; i < uLen; i++ {
		f.aData[i] = 0
	}
	f.aData[uLen] = uValue
}

// ToByteArray returns the byte representation identical to the Java code.
// If bNewVer is false, it writes bytes in a "reverse fill" big-endian order
// across the entire array (matching Java's decrementing uLen fill).
// If bNewVer is true, it writes each 32-bit word in little-endian order,
// from last word to first (ToByteArrayEx in Java).
func (f *Flag) ToByteArray(bNewVer bool) []byte {
	if f == nil {
		return nil
	}
	if bNewVer {
		return f.ToByteArrayEx()
	}
	uLen := len(f.aData) * 4
	pDest := make([]byte, uLen)

	idx := uLen
	for i := len(f.aData); i >= 1; i-- {
		uData := f.aData[i-1]
		idx--
		pDest[idx] = byte((uData >> 24) & 0xFF)
		idx--
		pDest[idx] = byte((uData >> 16) & 0xFF)
		idx--
		pDest[idx] = byte((uData >> 8) & 0xFF)
		idx--
		pDest[idx] = byte(uData & 0xFF)
	}
	return pDest
}

// ToByteArrayEx is the "reverse" of ToByteArray in Java: it outputs per-word
// little-endian bytes, iterating from the end towards the beginning.
func (f *Flag) ToByteArrayEx() []byte {
	if f == nil {
		return nil
	}
	pDest := make([]byte, len(f.aData)*4)
	uLen := 0
	for i := len(f.aData); i >= 1; i-- {
		uData := f.aData[i-1]
		pDest[uLen] = byte(uData & 0xFF)
		uLen++
		pDest[uLen] = byte((uData >> 8) & 0xFF)
		uLen++
		pDest[uLen] = byte((uData >> 16) & 0xFF)
		uLen++
		pDest[uLen] = byte((uData >> 24) & 0xFF)
		uLen++
	}
	return pDest
}

// ToHexString returns the hexadecimal representation, prefixed with 0x,
// concatenating each word as 8 uppercase hex digits (like Java).
func (f *Flag) ToHexString() string {
	if f == nil {
		return "0x"
	}
	var b strings.Builder
	b.WriteString("0x")
	for i := 0; i < len(f.aData); i++ {
		b.WriteString(fmt.Sprintf("%08X", f.aData[i]))
	}
	return b.String()
}

// ToHexBytes is a helper that renders the byte array in hex for debugging.
func (f *Flag) ToHexBytes(bNewVer bool) string {
	bs := f.ToByteArray(bNewVer)
	if bs == nil {
		return ""
	}
	return hex.EncodeToString(bs)
}
