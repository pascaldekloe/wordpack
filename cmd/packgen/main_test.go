package main_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

//go:generate go run . -package main_test gen_test.go

var randomYetConsistent = rand.New(rand.NewSource(42))

// TestDeltaPacks tests pack & unpack for each supported bit-size.
func TestDeltaPacks(t *testing.T) {
	for bitN := 0; bitN <= 64; bitN++ {
		t.Run(fmt.Sprintf("%dBitDelta", bitN), func(t *testing.T) {
			t.Run("uint64", func(t *testing.T) {
				testDeltaPack[uint64](t, bitN)
			})
			t.Run("int64", func(t *testing.T) {
				testDeltaPack[int64](t, bitN)
			})

			if bitN <= 32 {
				t.Run("uint32", func(t *testing.T) {
					testDeltaPack[uint32](t, bitN)
				})
				t.Run("int32", func(t *testing.T) {
					testDeltaPack[int32](t, bitN)
				})
			}
		})
	}
}

func testDeltaPack[T uint32 | int32 | uint64 | int64](t *testing.T, bitN int) {
	var data [64]T
	switch bitN {
	case 0:
		// same value causes zero delta
		x := T(randomYetConsistent.Uint64())
		for i := range data {
			data[i] = x
		}
	case 64:
		// no compression; delta equals word width
		for i := range data {
			data[i] = T(randomYetConsistent.Uint64())
		}
	default:
		// differ from previous value with up to bitN bits
		for i := range data {
			data[i] = T(randomYetConsistent.Int63n(int64(bitN)))
			if i != 0 {
				data[i] += data[i-1]
			}
		}
	}

	in := data // copy just in case pack mutates input
	pack := appendDeltaPackNBit(nil, &in, bitN, data[0])

	got, err := appendDeltaUnpackNBit(nil, pack, bitN, data[0])
	if err != nil {
		t.Logf("packed as: %#x", pack)
		t.Fatal("unpack error:", err)
	}
	want := data[:]
	if !reflect.DeepEqual(got, want) {
		t.Logf("packed as: %#x", pack)
		t.Errorf("pack + unpack changed input\ngot:  %#x\nwant: %#x", got, want)
	}
}

func BenchmarkDeltaBitPacks(b *testing.B) {
	for _, bitN := range []int{1, 7, 32, 63} {
		b.Run(fmt.Sprintf("%dBitDelta", bitN), func(b *testing.B) {
			b.Run("uint64", func(b *testing.B) {
				benchmarkDeltaBitPack[uint64](b, bitN)
			})
			b.Run("int64", func(b *testing.B) {
				benchmarkDeltaBitPack[int64](b, bitN)
			})

			if bitN <= 32 {
				b.Run("uint32", func(b *testing.B) {
					benchmarkDeltaBitPack[uint32](b, bitN)
				})
				b.Run("int32", func(b *testing.B) {
					benchmarkDeltaBitPack[int32](b, bitN)
				})
			}
		})
	}
}

func benchmarkDeltaBitPack[T uint64 | int64 | uint32 | int32](b *testing.B, bitN int) {
	var data [64]T
	for i := range data {
		data[i] = T(randomYetConsistent.Int63n(int64(bitN)))
		if i > 0 {
			data[i] += data[i-1]
		}
	}

	b.Run("Pack", func(b *testing.B) {
		var dst []uint64 // bufer reused
		for i := 0; i < b.N; i++ {
			dst = appendDeltaPackNBit(dst[:0], &data, bitN, data[0])
		}
	})

	b.Run("Unpack", func(b *testing.B) {
		src := appendDeltaPackNBit(nil, &data, bitN, data[0])

		var dst []T // buffer reused
		for i := 0; i < b.N; i++ {
			var err error
			dst, err = appendDeltaUnpackNBit(dst[:0], src, bitN, data[0])
			if err != nil {
				b.Fatal("unpack error:", err)
			}
		}
	})
}
