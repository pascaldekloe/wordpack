package pack64_test

import (
	"fmt"

	"github.com/pascaldekloe/wordpack/pack64"
)

func Example_codec() {
	var data = [64]int{
		99, 100, -101, 1001, 0, 0, 0, 1,
		99, 100, -101, 1001, 0, 0, 0, 2,
		99, 100, -101, 1001, 0, 0, 0, 3,
		99, 100, -101, 1001, 0, 0, 0, 4,
		99, 100, -101, 1001, 0, 0, 0, 5,
		99, 100, -101, 1001, 0, 0, 0, 6,
		99, 100, -101, 1001, 0, 0, 0, 7,
		99, 100, -101, 1001, 0, 0, 0, 8,
	}

	pack := pack64.AppendDeltaEncode(nil, &data, data[0])
	fmt.Printf("compressed %d integers into %d ✓\n", len(data), len(pack))

	var got [64]int
	pack64.AppendDeltaDecode(got[:0], pack, 99)
	if got == data {
		fmt.Println("got input back after codec cycle ✓")
	}
	// Output:
	// compressed 64 integers into 12 ✓
	// got input back after codec cycle ✓
}
