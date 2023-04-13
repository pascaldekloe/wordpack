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


## Credits

* Ronan Harmegnies — reference implementation
* Daniel Lemire — research
