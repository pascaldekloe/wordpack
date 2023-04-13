// Packacke pack64 provides compression for batches of 64 integers.
package pack64

import (
	"io"
	"reflect"
	"unsafe"
)

//go:generate go run ../cmd/packgen -limit 63 gen.go

// Write writes each Word marshalled in native endianness.
// The n return has the amount of bytes written—not words!
func Write(w io.Writer, words []Word) (n int, err error) {
	wordsHeader := (*reflect.SliceHeader)(unsafe.Pointer(&words))
	bytesHeader := *wordsHeader // copy
	bytesHeader.Len *= 64 / 8
	bytesHeader.Cap *= 64 / 8
	bytes := *(*[]byte)(unsafe.Pointer(&bytesHeader))
	n, err = w.Write(bytes)
	return n, err
}

// ReadFull reads exactly len(buf) Words from r into buf, unmarshalled in native
// endianness. The n return has the number of bytes read—not Words! The error is
// io.EOF only if no bytes were read. If an EOF happens after reading some but
// not all of the words, then ReadFull returns io.ErrUnexpectedEOF.
func ReadFull(r io.Reader, buf []Word) (n int, err error) {
	wordsHeader := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	bytesHeader := *wordsHeader // copy
	bytesHeader.Len *= 64 / 8
	bytesHeader.Cap *= 64 / 8
	bytes := *(*[]byte)(unsafe.Pointer(&bytesHeader))
	n, err = io.ReadFull(r, bytes)
	// zero remaining bytes of an incomplete word read, if any
	for i := n; i&7 != 0; i++ {
		bytes[i] = 0
	}
	return n, err
}
