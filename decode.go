package ber

// import (
//   "math/bits"
// )

func decodeIdentifier(b []byte) (uint8, uint8, uint32) {
	if len(b) == 0 {
		return 0, 0, 0
	}
	class, kind, tag := b[0]>>6, (b[0]>>5)&0x01, uint32(b[0]&0x1F)
	if tag == 0x1F {
		tag = decode128(b[1:])
	}
	return uint8(class), uint8(kind), tag
}

func decodeLength(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	if b[0]>>7 == 0 {
		return int(b[0] & 0x7F)
	}
	var (
		i int64
		c = int(b[0] & 0x7F)
	)
	for j := 0; j < c && j < len(b); j++ {
		i = (i << 8) | int64(b[j+1])
	}
	return int(i)
}

func decode128(b []byte) uint32 {
	if len(b) == 0 {
		return 0
	}
	var i uint32
	for j := 0; j < len(b); j++ {
		i = (i << 7) | uint32(b[j]&0x7F)
		if b[j]>>7 == 0 {
			break
		}
	}
	return i
}
