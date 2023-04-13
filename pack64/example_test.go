package pack64_test

import (
	"fmt"

	"github.com/pascaldekloe/wordpack/pack64"
)

// Full-cycle demonstration.
func Example_codec() {
	var data = [64]int{
		99, 100, -101, 1, 2, 3, 4, 5,
		144, 145, 146, 147, 148, 149, 150, 151,
		244, 245, 246, 247, 248, 249, 250, 251,
		344, 345, 346, 347, 348, 349, 350, 351,
		444, 445, 446, 447, 448, 449, 450, 451,
		544, 545, 546, 547, 548, 549, 550, 551,
		644, 645, 646, 647, 648, 649, 650, 651,
		744, 745, 746, 747, 748, 749, 750, 1001,
	}

	pack := pack64.AppendDeltaEncode(nil, &data, data[0])
	fmt.Printf("compressed %d integers into %d words ✓\n", len(data), len(pack))

	var got [64]int
	pack64.AppendDeltaDecode(got[:0], pack, 99)
	if got == data {
		fmt.Println("got input back after codec cycle ✓")
	}
	// Output:
	// compressed 64 integers into 9 words ✓
	// got input back after codec cycle ✓
}
