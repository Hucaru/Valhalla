package channel

// Flag is a bitset stored as 32-bit words (most significant bits at lower indices).
type Flag []uint32

// Version/collection bit lengths
const (
	Ver_1  = 32
	Ver_14 = 64

	Default = Ver_14 // Most common versions used are UINT128.
)

// NewFlag returns a flag with Default bits (64).
func NewFlag() *Flag {
	return NewFlagBits(Default)
}

// NewFlagBits constructs a Flag with exactly uBits capacity.
func NewFlagBits(uBits int) *Flag {
	if uBits < 0 {
		uBits = 0
	}
	words := uBits >> 5 // divide by 32
	f := Flag(make([]uint32, words))
	f.SetValue(0)
	return &f
}

// Data returns the internal data slice (do not modify).
func (f *Flag) Data() []uint32 {
	if f == nil {
		return nil
	}
	return []uint32(*f)
}

// IsZero reports whether all bits are zero (or the flag has no words).
func (f *Flag) IsZero() bool {
	if f == nil || len(*f) == 0 {
		return true
	}
	for _, w := range *f {
		if w != 0 {
			return false
		}
	}
	return true
}

// SetBitNumber sets a specific bit to 0 or 1.
// Bits are addressed MSB-first per 32-bit word.
func (f *Flag) SetBitNumber(uBit int, uValue int) {
	if f == nil || len(*f) == 0 {
		return
	}
	totalBits := 32 * len(*f)
	if uBit < 0 || uBit >= totalBits {
		return
	}
	word := uBit >> 5
	mask := uint32(1) << uint(31-(uBit&0x1F))
	(*f)[word] |= mask
	if uValue == 0 {
		(*f)[word] ^= mask
	}
}

// SetValue assigns a scalar 32-bit value to the least significant word,
// zeroing all higher words (mirrors Java).
func (f *Flag) SetValue(uValue uint32) {
	if f == nil || len(*f) == 0 {
		return
	}
	uLen := len(*f) - 1
	for i := 0; i < uLen; i++ {
		(*f)[i] = 0
	}
	(*f)[uLen] = uValue
}

// ToByteArray returns the byte representation identical to the Java code.
// If bNewVer is false, it writes bytes in a "reverse fill" big-endian order
// across the entire array (matching Java's decrementing uLen fill).
// If bNewVer is true, it writes each 32-bit word in little-endian order,
// from last word to first.
func (f *Flag) ToByteArray(bNewVer bool) []byte {
	if f == nil {
		return nil
	}
	if bNewVer {
		return f.ToByteArrayEx()
	}
	uLen := len(*f) * 4
	pDest := make([]byte, uLen)

	idx := uLen
	for i := len(*f); i >= 1; i-- {
		uData := (*f)[i-1]
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

// ToByteArrayEx is the "reverse" of ToByteArray: it outputs per-word
// little-endian bytes, iterating from the end towards the beginning.
func (f *Flag) ToByteArrayEx() []byte {
	if f == nil {
		return nil
	}
	pDest := make([]byte, len(*f)*4)
	uLen := 0
	for i := len(*f); i >= 1; i-- {
		uData := (*f)[i-1]
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
