package pack32

import (
	"math/bits"
	"testing"
)

func deltaDump(out *byte) // for debugging only

func TestIncrement(t *testing.T) {
	offset := int64(1001)
	var in [32]int64
	for i := range in {
		in[i] = offset + int64(i) + 1
	}
	t.Logf("input %#x", in)

	testPack(t, &in, offset)
}

func TestDecrement(t *testing.T) {
	offset := int64(1001)
	var in [32]int64
	for i := range in {
		in[i] = offset - int64(i) - 1
	}
	t.Logf("input %#x", in)

	testPack(t, &in, offset)
}

func testPack(t *testing.T, in *[32]int64, offset int64) {
	mask := delta64(in, offset)
	t.Logf("got mask %#x", mask)

	var deltas [32 * 8]byte
	deltaDump(&deltas[0])
	t.Logf("got deltas %#x", deltas)

	var out [32 * 8]byte
	switch bits.Len64(mask) {
	case 1:
		pack1(&out)
	case 2:
		pack2(&out)
	default:
		t.Fatal("wrong mask length")
	}
	t.Logf("got output %#x", out)
}

func BenchmarkPack64(b *testing.B) {
	b.SetBytes(32 * 8)
	offset := int64(1001)
	var in [32]int64
	for i := range in {
		in[i] = offset + int64(i) + 1
	}

	var out [32 * 8]byte
	for i := 0; i < b.N; i++ {
		n := DeltaEncode64(&out, &in, offset)
		if n != 4 {
			b.Fatalf("delta-pack wrote %d bytes, want 4", n)
		}
	}
}
