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

var (
	packageNameFlag = flag.String("package", "", "Overrides the detected package-`name`.")
	wordWidthFlag   = flag.Int("width", 64, "Sets the word size in `bits`.")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	// locate output file
	args := flag.Args()
	if len(args) != 1 {
		log.Fatal(os.Args[0] + ": need one output-file argument, and one only")
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
	c := Config{PackageName: *packageNameFlag, WordWidth: *wordWidthFlag}
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
}

// BitPacks returns each supported pack size in ascending order.
// Zero bits are not packed and neither is the word width itself.
func (c Config) BitPacks() []BitPack {
	packs := make([]BitPack, c.WordWidth-1)
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

// DeltaPackExpressions returns the expressions for each output word.
func (p BitPack) DeltaPackExpressions() []string {
	// input index
	var i int
	// number of input bits remaining from last word
	var passBitN int

	words := make([]string, 0, p.BitN)
	for range words[:p.BitN] {
		// expression text
		var buf bytes.Buffer

		// number of bits free in output word
		space := p.WordWidth

		if len(words) == 0 {
			fmt.Fprintf(&buf, "|uint%d(src[%d]-offset)<<%d", p.WordWidth, i, space-p.BitN)
			space -= p.BitN
			i++
		}
		if passBitN > 0 {
			fmt.Fprintf(&buf, "|uint%d(src[%d]-src[%d])<<%d", p.WordWidth, i, i-1, space-passBitN)
			space -= passBitN
			i++
			passBitN = 0
		}
		for ; space >= p.BitN; space -= p.BitN {
			fmt.Fprintf(&buf, "|uint%d(src[%d]-src[%d])<<%d", p.WordWidth, i, i-1, space-p.BitN)
			i++
		}
		if space > 0 {
			passBitN = p.BitN - space
			fmt.Fprintf(&buf, "|uint%d(src[%d]-src[%d])>>%d", p.WordWidth, i, i-1, passBitN)
		}

		words = append(words, strings.TrimPrefix(buf.String(), "|"))
	}

	return words
}

// UnpackExpressions returns the expressions for each output word.
func (p BitPack) UnpackExpressions() []string {
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
			words[i] = fmt.Sprintf("src[%d] >> %d & %#x", wordOffset, bitTailN, mask)
		} else {
			mask &^= 1<<(-bitTailN) - 1
			words[i] = fmt.Sprintf("(src[%d] << %d & %#x)", wordOffset, -bitTailN, mask)
			words[i] += fmt.Sprintf(" | (src[%d] >> %d)", wordOffset+1, p.WordWidth+bitTailN)
		}
	}
	return words
}
