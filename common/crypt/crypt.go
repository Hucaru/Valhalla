package crypt

func GetPacketLength(header []byte) int {
	var num = ((int32(header[0])) | (int32(header[1]) << 8)) | (int32(header[2]) << 16) | (int32(header[3]) << 24)
	return int((num >> 16) ^ (num & 0xFFFF))
}
