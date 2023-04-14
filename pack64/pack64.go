// Packacke pack64 provides compression for batches of 64 integers.
package pack64

import (
	"io"
	"unsafe"
)

//go:generate go run ../cmd/packgen -limit 63 gen.go

// Write writes each Word marshalled in native endianness.
// The n return has the amount of bytes written—not words!
func Write(w io.Writer, words []Word) (n int, err error) {
	p := (*byte)(unsafe.Pointer(unsafe.SliceData(words)))
	return w.Write(unsafe.Slice(p, len(words)*64/8))
}

// ReadFull reads exactly len(buf) Words from r into buf, unmarshalled in native
// endianness. The n return has the number of bytes read—not Words! The error is
// io.EOF only if no bytes were read. If an EOF happens after reading some but
// not all of the words, then ReadFull returns io.ErrUnexpectedEOF.
func ReadFull(r io.Reader, buf []Word) (n int, err error) {
	p := (*byte)(unsafe.Pointer(unsafe.SliceData(buf)))
	bytes := unsafe.Slice(p, len(buf)*64/8)
	n, err = io.ReadFull(r, bytes)
	// zero remaining bytes of an incomplete word read, if any
	for i := n; i&7 != 0; i++ {
		bytes[i] = 0
	}
	return n, err
}
