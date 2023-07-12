//go:build arm64 && !omitarm64

// Package pack32 provides compression for batches of 32 integers.
package pack32

import (
	"math/bits"
	"unsafe"
)

// Delta64 zig-zag encodes the difference of 32 consecutive values from src into
// vector registers, and it returns the bits in use by all encodings combined.
// The first value in src gets compared against offset.
func delta64(src *[32]int64, offset int64) (mask uint64)

// Pack1 takes 1 bit from each delta of delta64 which makes for 4 bytes of
// output. Pack2 takes 2 bits to write 8 bytes, and so forth until the 64 bit
// take with 256 output bytes.
func pack1(out *[32 * 8]byte)
func pack2(out *[32 * 8]byte)

// DeltaEncode64 returns the number of bytes written to dst (range 0–256).
func DeltaEncode64[Integer ~uint64 | ~int64](dst *[32 * 8]byte, src *[32]Integer, offset Integer) int {
	mask := delta64((*[32]int64)(unsafe.Pointer(src)), int64(offset))
	bitN := bits.Len64(mask)
	switch bitN {
	case 0:
		return 0
	case 1:
		pack1(dst)
	case 2:
		pack2(dst)
	default:
		panic(bitN)
	}
	return bitN << 2
}
