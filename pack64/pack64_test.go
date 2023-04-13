package pack64

import (
	"bytes"
	"fmt"
	"math/bits"
	"math/rand"
	"reflect"
	"testing"
)

// TestWordIO verifies a Write + ReadFull cycle.
func TestWordIO(t *testing.T) {
	// test data
	const (
		word0 = Word(0xaaaa_1111_2222_5555)
		word1 = Word(0x8866_4422_7755_3311)
	)

	// marshal to buffer
	var buf bytes.Buffer
	n, err := Write(&buf, []Word{word0, word1})
	if err != nil {
		t.Fatal("write error:", err)
	}
	if n != 16 {
		t.Errorf("wrote 2 words in %d bytes, want 16", n)
	}

	// unmarhal buffer
	got := make([]Word, 2)
	n, err = ReadFull(&buf, got)
	if err != nil {
		t.Error("read error:", err)
	}
	if n != 16 {
		t.Errorf("read %d bytes for 2 words, want 16", n)
	}

	if got[0] != word0 || got[1] != word1 {
		t.Errorf("read %#x, want [%#x, %#x]", got, word0, word1)
	}
}

// TestIncrementDelta verifies that an incrementing counter fits the single-bit
// pack.
func TestIncrementDelta(t *testing.T) {
	// test input with each value 1 more than the previous
	offset := int32(-10)
	var input [64]int32
	for i := range input {
		input[i] = offset + int32(i+1)
	}

	pack := append1BitDeltaEncode(nil, &input, offset)
	if len(pack) != 1 || pack[0] != 0xffff_ffff_ffff_ffff {
		t.Errorf("packed as %#x, want [0xffffffffffffffff]", pack)
	}

	got := append1BitDeltaDecode(nil, (*[1]Word)(pack), offset)
	for i := range got {
		if got[i] != input[i] {
			t.Errorf("encode + decode changed input word[%d]: got %#x, want %#x", i, got[i], input[i])
		}
	}
}

// TestDecrementDelta verifies that a decrementing counter fits the double-bit
// pack.
func TestDecrementDelta(t *testing.T) {
	// test input with each value 1 less than the previous
	offset := int64(10)
	var input [64]int64
	for i := range input {
		input[i] = offset - int64(i+1)
	}

	pack := append2BitDeltaEncode(nil, &input, offset)
	if len(pack) != 2 || pack[0] != 0xaaaa_aaaa_aaaa_aaaa || pack[1] != 0xaaaa_aaaa_aaaa_aaaa {
		t.Errorf("packed as %#x, want [0xaaaaaaaaaaaaaaaa 0xaaaaaaaaaaaaaaaa]", pack)
	}

	got := append2BitDeltaDecode(nil, (*[2]Word)(pack), offset)
	for i := range got {
		if got[i] != input[i] {
			t.Errorf("encode + decode changed input word[%d]: got %#x, want %#x", i, got[i], input[i])
		}
	}
}

// TestDeltaEncoding tests encode & decode for each supported bit-size.
func TestDeltaEncoding(t *testing.T) {
	for bitN := 0; bitN <= 64; bitN++ {
		t.Run(fmt.Sprintf("%dBitDelta", bitN), func(t *testing.T) {
			if bitN <= 16 {
				t.Run("int16", func(t *testing.T) {
					testDeltaEncoding[int16](t, bitN)
				})
			}
			if bitN <= 32 {
				t.Run("int32", func(t *testing.T) {
					testDeltaEncoding[int32](t, bitN)
				})
			}
			t.Run("int64", func(t *testing.T) {
				testDeltaEncoding[int64](t, bitN)
			})
			t.Run("uint64", func(t *testing.T) {
				testDeltaEncoding[uint64](t, bitN)
			})
		})
	}
}

func testDeltaEncoding[T Integer](t *testing.T, bitN int) {
	data, offset := randomNBitDeltas[T](t, bitN)

	in := data // copy just in case encode mutates input
	pack := AppendDeltaEncode(nil, &in, offset)

	if len(pack) != bitN {
		t.Errorf("packed %d-bit random data in %d words, want %d", bitN, len(pack), bitN)
	}

	got := AppendDeltaDecode(nil, pack, offset)
	want := data[:]
	if !reflect.DeepEqual(got, want) {
		t.Logf("packed as: %#x", pack)
		t.Errorf("encode + decode changed input\ngot:  %#x\nwant: %#x", got, want)
	}
}

func BenchmarkDeltaBitEncoding(b *testing.B) {
	for _, bitN := range []int{1, 7, 32, 63} {
		b.Run(fmt.Sprintf("%dBitDelta", bitN), func(b *testing.B) {
			if bitN <= 16 {
				b.Run("int16", func(b *testing.B) {
					benchmarkDeltaBitEncoding[int16](b, bitN)
				})
			}
			if bitN <= 32 {
				b.Run("int32", func(b *testing.B) {
					benchmarkDeltaBitEncoding[int32](b, bitN)
				})
			}
			b.Run("int64", func(b *testing.B) {
				benchmarkDeltaBitEncoding[int64](b, bitN)
			})
			b.Run("uint64", func(b *testing.B) {
				benchmarkDeltaBitEncoding[uint64](b, bitN)
			})
		})
	}
}

func benchmarkDeltaBitEncoding[T Integer](b *testing.B, bitN int) {
	data, offset := randomNBitDeltas[T](b, bitN)

	b.Run("Encode", func(b *testing.B) {
		var dst []Word // buffer reused
		for i := 0; i < b.N; i++ {
			dst = AppendDeltaEncode(dst[:0], &data, offset)
		}

		b.StopTimer()
		b.ReportMetric(float64(b.N*64)/b.Elapsed().Seconds(), "ℕ/s")
	})

	b.Run("Decode", func(b *testing.B) {
		src := AppendDeltaEncode(nil, &data, offset)

		var dst []T // buffer reused
		for i := 0; i < b.N; i++ {
			dst = AppendDeltaDecode(dst[:0], src, offset)
		}

		b.StopTimer()
		b.ReportMetric(float64(b.N*64)/b.Elapsed().Seconds(), "ℕ/s")
	})
}

// RandomNBitDelta generates a pseudo random data set with it's deltas zig-zag
// encoded less than or equal to bitN in size.
func randomNBitDeltas[T Integer](t testing.TB, bitN int) (data [64]T, offset T) {
	randomYetConsistent := rand.New(rand.NewSource(42))

	offset = T(randomYetConsistent.Uint64())

	switch bitN {
	case 0:
		// same value causes zero delta
		for i := range data {
			data[i] = offset
		}
		return

	case bits.OnesCount64(uint64(^T(0))):
		// no compression; delta equals word width
		for i := range data {
			data[i] = T(randomYetConsistent.Uint64())
		}
		return
	}

	// limit bit size of (zig-zag encoded) deltas
	mask := T(1)<<bitN - 1

	for i := range data {
		pass := offset
		if i > 0 {
			pass = data[i-1]
		}

		for {
			zigZagDelta := randomYetConsistent.Int63() & int64(mask)
			// decode
			delta := (zigZagDelta >> 1) ^ -(zigZagDelta & 1)
			// apply
			data[i] = pass - T(delta)
			// overflow check
			if (data[i] < pass) != (delta > 0) || (data[i] > pass) != (delta < 0) {
				t.Logf("retry on random delta %#x (zig-zag encodes as %#x) as it overflows %T offset %#x",
					delta, zigZagDelta, offset, offset)
				continue
			}
			break
		}
	}

	return
}
