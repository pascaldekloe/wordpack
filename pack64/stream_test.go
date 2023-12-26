package pack64

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestStreamNone(t *testing.T) {
	const deltaOffset uint64 = 0

	var buf bytes.Buffer
	err := NewWriter(&buf, deltaOffset).Flush(nil)
	if err != nil {
		t.Fatal("flush error:", err)
	}

	if l := buf.Len(); l != 0 {
		t.Fatalf("wrote %d bytes, want none", l)
	}

	got, err := NewReader(&buf, deltaOffset).ReadAppend(nil)
	if err != io.EOF {
		if err != nil {
			t.Errorf("got read error %s, want EOF", err)
		} else {
			t.Errorf("read got %d, want EOF", got)
		}
	}
}

func TestStream(t *testing.T) {
	// test values
	const deltaOffset uint64 = 42
	data := make([]uint64, PageSize)
	for i := range data {
		data[i] = uint64(i) + deltaOffset
	}

	for n := 1; n <= len(data); n++ {
		feed := data[:n]

		var buf bytes.Buffer
		err := NewWriter(&buf, deltaOffset).Flush(feed)
		if err != nil {
			t.Fatalf("flush with %d numbers got error: %s", n, err)
		}

		wantSize := 8            // 1 header
		wantSize += (n / 64) * 8 // increments need 1 bit each
		wantSize += (n % 64) * 8 // uncompressed tail
		if l := buf.Len(); l != wantSize {
			t.Errorf("flush with %d number encoded to %d bytes, want %d bytes", n, l, wantSize)
		}

		r := NewReader(&buf, deltaOffset)
		var got []uint64
		for err == nil {
			got, err = r.ReadAppend(got)
		}
		if err != io.EOF {
			t.Fatalf("stream with %d numbers got read error: %s", n, err)
		}
		if l := buf.Len(); l != 0 {
			t.Errorf("%d bytes remaining after read in stream with %d numbers", l, n)
		}
		if !reflect.DeepEqual(got, feed) {
			t.Fatalf("wrote %d to stream, read %d back", feed, got)
		}
	}
}

func BenchmarkWritePack(b *testing.B) {
	deltaOffset := int64(b.N)
	w := NewWriter(io.Discard, deltaOffset)

	// test data
	var p [64]int64
	for i := range p {
		p[i] = int64(i&3) + deltaOffset
	}

	b.SetBytes(64 * 8)
	for i := 0; i < b.N; i++ {
		err := w.WritePack(&p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadAppend(b *testing.B) {
	deltaOffset := int64(b.N)

	// test data
	var p [PageSize]int64
	for i := range p {
		p[i] = int64(i&3) + deltaOffset
	}

	var buf bytes.Buffer
	err := NewWriter(&buf, deltaOffset).Flush(p[:])
	if err != nil {
		b.Fatal(err)
	}
	stream := bytes.Repeat(buf.Bytes(), 1024)
	streamPackN := 1024 * PageSize / 64

	b.SetBytes(64 * 8)
	b.ResetTimer()
	var r *Reader[int64]
	for i := 0; i < b.N; i++ {
		if i%streamPackN == 0 {
			r = NewReader(bytes.NewReader(stream), deltaOffset)
		}
		got, err := r.ReadAppend(p[:0])
		if err != nil {
			b.Fatal("read error:", err)
		}
		if len(got) != 64 {
			b.Fatalf("read %d numbers, want 64", len(got))
		}
	}
}
