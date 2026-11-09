// Command gowordtext prints the text goword extracts from a .docx file and can
// write a goword round-trip copy. It is the library side of the differential
// harness in scripts/diffharness.py.
package main

import (
	"fmt"
	"os"

	goword "github.com/VantageDataChat/GoWord"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: gowordtext extract <in.docx> | gowordtext roundtrip <in.docx> <out.docx>")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "extract":
		doc, err := goword.Open(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(doc.ExtractText())
	case "roundtrip":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: gowordtext roundtrip <in.docx> <out.docx>")
			os.Exit(2)
		}
		doc, err := goword.Open(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := doc.Save(os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(doc.ExtractText())
	default:
		fmt.Fprintln(os.Stderr, "unknown command")
		os.Exit(2)
	}
}
