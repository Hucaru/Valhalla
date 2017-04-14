package crypt

// GetPacketLength -
func GetPacketLength(header []byte) int {
	var num = ((int32(header[0])) | (int32(header[1]) << 8)) | (int32(header[2]) << 16) | (int32(header[3]) << 24)
	return int((num >> 16) ^ (num & 0xFFFF))
}

func ror(val byte, num int) byte {
	for i := 0; i < num; i++ {
		var lowbit int

		if val&1 > 0 {
			lowbit = 1
		} else {
			lowbit = 0
		}

		val >>= 1
		val |= byte(lowbit << 7)
	}

	return val
}

func rol(val byte, num int) byte {
	var highbit int

	for i := 0; i < num; i++ {
		if val&0x80 > 0 {
			highbit = 1
		} else {
			highbit = 0
		}

		val <<= 1
		val |= byte(highbit)
	}

	return val
}

// Decrypt - Taken from Kogami
func Decrypt(buf []byte) {
	var j int32
	var a, b, c byte

	for i := byte(0); i < 3; i++ {
		a = 0
		b = 0

		for j = int32(len(buf)); j > 0; j-- {
			c = buf[j-1]
			c = rol(c, 3)
			c ^= 0x13
			a = c
			c ^= b
			c = byte(int32(c) - j)
			c = ror(c, 4)
			b = a
			buf[j-1] = c
		}

		a = 0
		b = 0

		for j = int32(len(buf)); j > 0; j-- {
			c = buf[int32(len(buf))-j]
			c -= 0x48
			c ^= 0xFF
			c = rol(c, int(j))
			a = c
			c ^= b
			c = byte(int32(c) - j)
			c = ror(c, 3)
			b = a
			buf[int32(len(buf))-j] = c
		}
	}
}
