//go:build arm64 && !omitarm64

package pack32

import (
	"fmt"
	"math/bits"
	"testing"
)

func deltaDump(out *byte) // for debugging only

const testOffset = int64(1001)

func TestIncrements(t *testing.T) {
	for _, incN := range []int64{1, 2, 3} {
		var in [32]int64
		for i := range in {
			in[i] = testOffset + int64(i+1)*incN
		}
		t.Logf("input %#x", in)
		testPack(t, &in, testOffset)
	}
}

func TestDecrements(t *testing.T) {
	for _, decN := range []int64{1, 2} {
		var in [32]int64
		for i := range in {
			in[i] = testOffset - int64(i+1)*decN
		}
		t.Logf("input %#x", in)
		testPack(t, &in, testOffset)
	}
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
		pack1bit(&out)
	case 2:
		pack2bit(&out)
	case 3:
		pack3bit(&out)
	default:
		t.Fatal("wrong mask length")
	}
	t.Logf("got output %#x", out)
}

func BenchmarkPack64(b *testing.B) {
	for _, bitN := range []int64{1, 2, 3} {
		b.Run(fmt.Sprintf("%d-bit", bitN), func(b *testing.B) {
			b.SetBytes(32 * 8)
			var in [32]int64
			for i := range in {
				in[i] = testOffset + int64(i+1)*bitN
			}

			wantByteN := int(bitN * 4)

			var out [32 * 8]byte
			for i := 0; i < b.N; i++ {
				n := DeltaEncode64(&out, &in, testOffset)
				if n != wantByteN {
					b.Fatalf("delta-pack wrote %d bytes, want %d", n, wantByteN)
				}
			}
			b.ReportMetric(float64(b.N*32/1E9)/b.Elapsed().Seconds(), "Gℕ/s")
		})
	}
}
