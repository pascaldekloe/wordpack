// Package pack64 provides compression for batches of 64 integers.
package pack64

import "io"

// PageSize is the frame capacity for streams from Writer.
const PageSize = 9 * 64

// Writer encodes 64-bit integers to a stream.
type Writer[T uint64 | int64] struct {
	out         io.Writer
	lastValue   T
	header      Word
	headerShift uint
	buf         []Word             // pending write
	mem         [PageSize + 1]Word // buf space
}

// NewWriter begins the stream with a user defined delta offset. Readers of the
// stream must use the exact same value to decode. Try to get close to the first
// integer written, or use zero (0) for unknown.
func NewWriter[T uint64 | int64](out io.Writer, deltaOffset T) *Writer[T] {
	w := &Writer[T]{
		out:       out,
		lastValue: deltaOffset,
	}
	w.buf = w.mem[:1]
	return w
}

// WritePack adds 64 integers to the stream. Errors only come from the output
// io.Writer. The Writer is left in an undefined state after error encounters.
func (w *Writer[T]) WritePack(p *[64]T) (fatal error) {
	start := len(w.buf)
	w.buf = AppendDeltaEncode(w.buf, p, w.lastValue)
	w.lastValue = p[63]

	// add encoding size (range 0..64) to page header
	w.header |= Word(len(w.buf)-start) << w.headerShift
	w.headerShift += 7

	if w.headerShift < 63 {
		return nil // partial page pending
	}
	w.header |= 1 << 63 // full-page flag

	// redundant check omits Go panic
	if len(w.buf) != 0 {
		// write page with header
		w.buf[0] = w.header
		_, fatal = Write(w.out, w.buf)

		// start over
		w.header, w.headerShift = 0, 0
		// reserve header location
		w.buf = w.buf[:1]
	}

	return
}

// Flush writes any and all pending data including p [optional] to the stream.
// Encoding is suboptimal when the total number of integers written (since the
// stream start or a previous Flush) is not a multiple of PageSize.
func (w *Writer[T]) Flush(p []T) error {
	// consume full packs
	for len(p) > 63 {
		err := w.WritePack((*[64]T)(p))
		if err != nil {
			return err
		}

		p = p[64:]
	}
	if len(p) == 0 && w.headerShift == 0 {
		// finished on complete page
		return nil
	}

	// incomplete pack gets no compression
	for _, v := range p {
		w.buf = append(w.buf, Word(v))
	}
	// A partial page uses its last pack-size to
	// count the number of integers that follow.
	w.header |= Word(len(p)) << 56
	// mark unused pack-sizes
	for w.headerShift < 56 {
		w.header |= 127 << w.headerShift
		w.headerShift += 7
	}
	// install with redundant check to omit Go panic
	if len(w.buf) != 0 {
		w.buf[0] = w.header
	}

	_, err := Write(w.out, w.buf)
	if err != nil {
		return err
	}

	// start over
	w.header, w.headerShift = 0, 0
	// redundant check to omit Go panic
	if len(w.buf) != 0 {
		// reserve header location
		w.buf = w.buf[:1]
	}
	return nil
}

// Reader decodes 64-bit integers from a stream.
type Reader[T uint64 | int64] struct {
	in        io.Reader // data source
	lastValue T         // delta offset for next pack
	// packs 9 size of 7 bis each plus a "full" flag
	header Word
	// position of next size in header is multiple of 7
	headerShift uint

	// read buffer equals .buf[.offset:.byteN/8]
	buf [PageSize + 1]Word
	// byte [!] count in buffer
	byteN int
	// index of buffer position
	offset int
}

// NewReader begins the stream with a user defined delta offset. The value must
// match the NewWriter used to create this stream.
func NewReader[T uint64 | int64](in io.Reader, deltaOffset T) *Reader[T] {
	return &Reader[T]{
		in:          in,
		lastValue:   deltaOffset,
		headerShift: 63, // start exhausted
	}
}

func (r *Reader[T]) ensureNWords(min int) error {
	for r.byteN/8-r.offset < min {
		// move remainder to buffer start
		if r.offset != 0 {
			r.byteN -= r.offset * 8
			copy(r.buf[:(r.byteN+7)/8], r.buf[r.offset:])
			r.offset = 0
		}

		n, err := ReadAsOf(r.in, r.buf[:], r.byteN)
		r.byteN += n
		if err != nil {
			return err
		}
	}

	return nil
}

// ReadAppend appends integers from the stream to dst, and it returns the
// extended buffer. Errors only come from the input io.Reader. The return
// equals dst when Read encounters an error. Otherwise, Reads are of at
// most PageSize integers in size.
func (r *Reader[T]) ReadAppend(dst []T) ([]T, error) {
	// need next header?
	if r.headerShift > 56 {
		err := r.ensureNWords(1)
		if err != nil {
			return dst, err
		}

		r.header = r.buf[r.offset]
		r.offset++
		r.headerShift = 0
	}

	size := int(r.header>>r.headerShift) & 127

	if r.header&(1<<63) == 0 && (r.headerShift == 56 || size == 127) {
		// incomplete pack in partial page
		remain := int(r.header >> 56)
		err := r.ensureNWords(remain)
		if err != nil {
			return dst, err
		}

		// copy without compression
		for _, w := range r.buf[r.offset : r.offset+remain] {
			dst = append(dst, T(w))
		}
		r.offset += remain
		r.headerShift = 63
		return dst, nil
	}

	err := r.ensureNWords(size)
	if err != nil {
		return dst, err
	}

	r.headerShift += 7
	enc := r.buf[r.offset : r.offset+size]
	r.offset += size

	dst = AppendDeltaDecode(dst, enc, r.lastValue)
	// redundant check omits Go panic
	if len(dst) != 0 {
		r.lastValue = dst[len(dst)-1]
	}
	return dst, nil
}
