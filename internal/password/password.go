package password

import (
	"encoding/binary"
	"strings"
)

const (
	alphabet    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	blockOffset = 0x3E8
	blockSize   = 4
	blockDigits = 7
)

var bitOrder = [32]uint8{
	26, 31, 17, 10, 30, 16, 24, 2,
	29, 8, 20, 15, 28, 11, 13, 4,
	19, 23, 0, 12, 14, 27, 6, 18,
	21, 3, 9, 7, 22, 1, 25, 5,
}

func Encode(plain string) string {
	src := []byte(plain)
	blocks := (len(src) + blockSize - 1) / blockSize

	var out strings.Builder
	out.Grow(blocks * blockDigits)
	for i := range blocks {
		var block [blockSize]byte
		copy(block[:], src[i*blockSize:])
		value := binary.LittleEndian.Uint32(block[:]) + blockOffset
		writeBase36(&out, shuffle(value))
	}
	return out.String()
}

func shuffle(v uint32) uint32 {
	var out uint32
	for i := 0; v != 0; i++ {
		if v&1 != 0 {
			out += 1 << bitOrder[i]
		}
		v >>= 1
	}
	return out
}

func writeBase36(out *strings.Builder, v uint32) {
	for range blockDigits {
		out.WriteByte(alphabet[v%36])
		v /= 36
	}
}
