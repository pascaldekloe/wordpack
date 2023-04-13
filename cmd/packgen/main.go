// Package main generates code. See the manual for details.
package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Name the command in use.
var name = os.Args[0]

var (
	packageNameFlag = flag.String("package", "", "Overrides the `name` detected by default.")
	wordWidthFlag   = flag.Int("width", 64, "Sets the word size in `bits`.")
	packLimitFlag   = flag.Int("limit", 42, "Sets the upper boundary for bit-packing in `bits`. Full range\ncompression can be achieved with -limit set to one less than the\n-width value. Higher limits generate more code.")
)

func main() {
	log.SetFlags(0)
	flag.Usage = printManual
	flag.Parse()

	// locate output file
	args := flag.Args()
	switch len(args) {
	case 0:
		printManual()
		os.Exit(2)
		return
	case 1:
		break // OK
	default:
		log.Print(os.Args[0] + ": need one output-file argument only")
		os.Exit(2)
	}
	path := args[0]

	// unsure parent directory
	dir := filepath.Dir(path)
	if dir == "." {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := os.MkdirAll(dir, 0o777)
		if err != nil {
			log.Fatal(err)
		}
	}

	// open output file
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// execute configuration
	c := Config{
		PackageName: *packageNameFlag,
		WordWidth:   *wordWidthFlag,
		PackLimit:   *packLimitFlag,
	}
	if c.PackageName == "" {
		c.PackageName = filepath.Base(dir)
	}
	err = generatePack(f, c)
	if err != nil {
		log.Fatal(err)
	}
}

//go:embed pack.template
var packText string

func generatePack(w io.Writer, c Config) error {
	t := template.New("pack").Funcs(map[string]any{
		"iterate": func(n int) []int {
			all := make([]int, n)
			for i := range all {
				all[i] = i + 1
			}
			return all
		},
	})

	t, err := t.Parse(packText)
	if err != nil {
		return err
	}

	return t.Execute(w, c)
}

type Config struct {
	PackageName string
	WordWidth   int
	PackLimit   int
}

// BitPacks returns each supported pack size in ascending order.
// Zero bits are not packed and neither is the word width itself.
func (c Config) BitPacks() []BitPack {
	packN := c.WordWidth - 1
	if c.PackLimit >= 0 && packN > c.PackLimit {
		packN = c.PackLimit
	}
	packs := make([]BitPack, packN)
	for i := range packs {
		packs[i].BitN = i + 1
		packs[i].WordWidth = c.WordWidth
	}
	return packs
}

type BitPack struct {
	BitN      int
	WordWidth int
}

// BitPackExpressions returns Go code for each output word.
func (p BitPack) BitPackExpressions(inputExpressions []string) []string {
	// number of input bits remaining from last word
	var passBitN int

	words := make([]string, p.BitN)
	for i := range words {
		// expression text
		var buf bytes.Buffer

		// number of bits free in output word
		space := p.WordWidth

		if passBitN > 0 {
			fmt.Fprintf(&buf, "|(Word(%s)<<%d)", inputExpressions[0], space-passBitN)
			space -= passBitN
			inputExpressions = inputExpressions[1:]
			passBitN = 0
		}
		for ; space >= p.BitN; space -= p.BitN {
			fmt.Fprintf(&buf, "|(Word(%s)<<%d)", inputExpressions[0], space-p.BitN)
			inputExpressions = inputExpressions[1:]
		}
		if space > 0 {
			passBitN = p.BitN - space
			fmt.Fprintf(&buf, "|(Word(%s)>>%d)", inputExpressions[0], passBitN)
		}

		words[i] = strings.TrimPrefix(buf.String(), "|")
	}

	return words
}

// DeltaEncodeExpressions returns Go code for each input word, which calculates
// the zig-zag encoding of the differenences between each input word, and with
// "offset" for the input before src[0].
func (p BitPack) DeltaEncodeExpressions() []string {
	words := make([]string, p.WordWidth)
	for i := range words {
		var delta string
		if i == 0 {
			delta = fmt.Sprintf("int%d(offset-src[0])", p.WordWidth)
		} else {
			delta = fmt.Sprintf("int%d(src[%d]-src[%d])", p.WordWidth, i-1, i)
		}
		words[i] = fmt.Sprintf("(%s>>%d)^(%s<<1)", delta, p.WordWidth-1, delta)
	}
	return words
}

// BitUnpackExpressions returns the Go code for each encoded value.
func (p BitPack) BitUnpackExpressions() []string {
	words := make([]string, p.WordWidth)
	for i := range words {
		// calculate location in word input
		bitOffset := i * p.BitN
		wordOffset := bitOffset / p.WordWidth
		// caculate position in input word
		bitSkipN := bitOffset % p.WordWidth
		bitTailN := p.WordWidth - bitSkipN - p.BitN
		// negative bitTailN implies overflow to next word

		mask := 1<<p.BitN - 1
		if bitTailN >= 0 {
			words[i] = fmt.Sprintf("(src[%d]>>%d)&%#x", wordOffset, bitTailN, mask)
		} else {
			mask &^= 1<<(-bitTailN) - 1
			words[i] = fmt.Sprintf("((src[%d]<<%d)&%#x)", wordOffset, -bitTailN, mask)
			words[i] += fmt.Sprintf("|(src[%d]>>%d)", wordOffset+1, p.WordWidth+bitTailN)
		}
	}
	return words
}

// ANSI escape codes for markup
const (
	bold  = "\x1b[1m"
	clear = "\x1b[0m"
)

func printManual() {
	w := os.Stderr

	w.WriteString("NAME\n" +
		"\t" + name + clear + " — generate bit-pack code\n" +
		"\n" +
		bold + "SYNOPSIS\n" +
		"\t" + name + clear + " [" + bold + "OPTIONS" + clear + "] FILE\n" +
		"\n" +
		bold + "OPTIONS" + clear + "\n")

	flag.CommandLine.SetOutput(w)
	flag.CommandLine.PrintDefaults()

	w.WriteString("\n" +
		bold + "BUGS" + clear + "\n" +
		"\tReport bugs at <https://github.com/pascaldekloe/wordpack/issues>.\n")
}
