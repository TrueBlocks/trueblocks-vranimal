// vrml-fmt parses a VRML97 .wrl file and pretty-prints it to stdout.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/writer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: vrml-fmt <file.wrl>\n")
		os.Exit(1)
	}

	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	p := parser.NewParser(f)
	p.SetBaseDir(filepath.Dir(filename))
	nodes := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors:\n  %s\n", strings.Join(errs, "\n  "))
		os.Exit(1)
	}

	w := writer.New(os.Stdout)
	w.WriteScene(nodes)
}
