# WordPack

Delta compression with bit-packing as a library or as generated code.

This is free and unencumbered software released into the
[public domain](http://creativecommons.org/publicdomain/zero/1.0).

[![Build](https://github.com/pascaldekloe/wordpack/actions/workflows/go.yml/badge.svg)](https://github.com/pascaldekloe/wordpack/actions/workflows/go.yml)


## Library

The `pack64` directory provides compression for batches of 64 integers.

[![Go Reference](https://pkg.go.dev/badge/github.com/pascaldekloe/wordpack.svg)](https://pkg.go.dev/github.com/pascaldekloe/wordpack)


## Code Generator

Build your own code with the packgen(1) command.

```
NAME
	packgen — generate bit-pack code

SYNOPSIS
	packgen [OPTIONS] FILE

OPTIONS
  -limit bits
    	Sets the upper boundary for bit-packing in bits. Full range
    	compression can be achieved with -limit set to one less than the
    	-width value. Higher limits generate more code. (default 42)
  -package name
    	Overrides the name detected by default.
  -width bits
    	Sets the word size in bits. (default 64)

BUGS
	Report bugs at <https://github.com/pascaldekloe/wordpack/issues>.
```


### Benchmarks

The “billions of integers per second” are on par with
[other efforts](https://lemire.me/blog/2012/09/12/fast-integer-compression-decoding-billions-of-integers-per-second/).

```
DeltaBitEncoding/1BitDelta/int16/Encode-8     58.53n ± 0%
DeltaBitEncoding/1BitDelta/int32/Encode-8     59.73n ± 2%
DeltaBitEncoding/1BitDelta/int64/Encode-8     63.64n ± 1%
DeltaBitEncoding/1BitDelta/uint64/Encode-8    64.37n ± 1%
DeltaBitEncoding/7BitDelta/int16/Encode-8     49.45n ± 0%
DeltaBitEncoding/7BitDelta/int32/Encode-8     49.38n ± 0%
DeltaBitEncoding/7BitDelta/int64/Encode-8     42.24n ± 0%
DeltaBitEncoding/7BitDelta/uint64/Encode-8    42.20n ± 0%
DeltaBitEncoding/32BitDelta/int32/Encode-8    45.68n ± 0%
DeltaBitEncoding/32BitDelta/int64/Encode-8    39.72n ± 0%
DeltaBitEncoding/32BitDelta/uint64/Encode-8   39.64n ± 0%
DeltaBitEncoding/63BitDelta/int64/Encode-8    53.14n ± 0%
DeltaBitEncoding/63BitDelta/uint64/Encode-8   53.01n ± 0%

                                            │     ℕ/s     │
DeltaBitEncoding/1BitDelta/int16/Encode-8     1.093G ± 0%
DeltaBitEncoding/1BitDelta/int32/Encode-8     1.071G ± 2%
DeltaBitEncoding/1BitDelta/int64/Encode-8     1.006G ± 1%
DeltaBitEncoding/1BitDelta/uint64/Encode-8    994.2M ± 1%
DeltaBitEncoding/7BitDelta/int16/Encode-8     1.294G ± 0%
DeltaBitEncoding/7BitDelta/int32/Encode-8     1.296G ± 0%
DeltaBitEncoding/7BitDelta/int64/Encode-8     1.515G ± 0%
DeltaBitEncoding/7BitDelta/uint64/Encode-8    1.517G ± 0%
DeltaBitEncoding/32BitDelta/int32/Encode-8    1.401G ± 0%
DeltaBitEncoding/32BitDelta/int64/Encode-8    1.611G ± 0%
DeltaBitEncoding/32BitDelta/uint64/Encode-8   1.615G ± 0%
DeltaBitEncoding/63BitDelta/int64/Encode-8    1.204G ± 0%
DeltaBitEncoding/63BitDelta/uint64/Encode-8   1.207G ± 0%
```

```
DeltaBitEncoding/1BitDelta/int16/Decode-8     37.16n ± 1%
DeltaBitEncoding/1BitDelta/int32/Decode-8     36.96n ± 0%
DeltaBitEncoding/1BitDelta/int64/Decode-8     36.99n ± 0%
DeltaBitEncoding/1BitDelta/uint64/Decode-8    37.00n ± 0%
DeltaBitEncoding/7BitDelta/int16/Decode-8     37.19n ± 0%
DeltaBitEncoding/7BitDelta/int32/Decode-8     37.02n ± 0%
DeltaBitEncoding/7BitDelta/int64/Decode-8     37.02n ± 0%
DeltaBitEncoding/7BitDelta/uint64/Decode-8    37.02n ± 0%
DeltaBitEncoding/32BitDelta/int32/Decode-8    38.24n ± 0%
DeltaBitEncoding/32BitDelta/int64/Decode-8    38.23n ± 0%
DeltaBitEncoding/32BitDelta/uint64/Decode-8   38.23n ± 0%
DeltaBitEncoding/63BitDelta/int64/Decode-8    43.26n ± 0%
DeltaBitEncoding/63BitDelta/uint64/Decode-8   43.26n ± 0%

                                            │     ℕ/s     │
DeltaBitEncoding/1BitDelta/int16/Decode-8     1.723G ± 1%
DeltaBitEncoding/1BitDelta/int32/Decode-8     1.732G ± 0%
DeltaBitEncoding/1BitDelta/int64/Decode-8     1.730G ± 0%
DeltaBitEncoding/1BitDelta/uint64/Decode-8    1.729G ± 0%
DeltaBitEncoding/7BitDelta/int16/Decode-8     1.721G ± 0%
DeltaBitEncoding/7BitDelta/int32/Decode-8     1.729G ± 0%
DeltaBitEncoding/7BitDelta/int64/Decode-8     1.729G ± 0%
DeltaBitEncoding/7BitDelta/uint64/Decode-8    1.729G ± 0%
DeltaBitEncoding/32BitDelta/int32/Decode-8    1.674G ± 0%
DeltaBitEncoding/32BitDelta/int64/Decode-8    1.674G ± 0%
DeltaBitEncoding/32BitDelta/uint64/Decode-8   1.674G ± 0%
DeltaBitEncoding/63BitDelta/int64/Decode-8    1.479G ± 0%
DeltaBitEncoding/63BitDelta/uint64/Decode-8   1.479G ± 0%
```


## Credits

* Ronan Harmegnies — reference implementation
* Daniel Lemire — research
