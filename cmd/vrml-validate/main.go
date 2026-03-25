package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/validator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: vrml-validate <file.wrl> [file2.wrl ...]\n")
		os.Exit(1)
	}

	exitCode := 0
	for _, path := range os.Args[1:] {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			exitCode = 1
			continue
		}

		p := parser.NewParser(f)
		p.SetBaseDir(filepath.Dir(path))
		nodes := p.Parse()
		f.Close()

		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "%s: parse error: %s\n", path, e)
			}
			exitCode = 1
		}

		v := validator.New()
		findings := v.Validate(nodes)

		if len(findings) == 0 {
			fmt.Printf("%s: valid\n", path)
		} else {
			for _, finding := range findings {
				fmt.Printf("%s: %s\n", path, finding)
			}
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}
