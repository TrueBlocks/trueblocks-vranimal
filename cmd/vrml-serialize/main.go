package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/serializer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: vrml-serialize <file.wrl|file.vrb>\n")
		fmt.Fprintf(os.Stderr, "  .wrl -> .vrb  encode (VRML text to binary)\n")
		fmt.Fprintf(os.Stderr, "  .vrb -> .wrl  decode (binary to VRML text, via writer)\n")
		os.Exit(1)
	}

	path := os.Args[1]
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".wrl":
		if err := encodeFile(path); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case ".vrb":
		if err := decodeFile(path); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown extension %q (use .wrl or .vrb)\n", ext)
		os.Exit(1)
	}
}

func encodeFile(wrlPath string) error {
	f, err := os.Open(wrlPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close error: %v\n", err)
		}
	}()

	p := parser.NewParser(f)
	p.SetBaseDir(filepath.Dir(wrlPath))
	nodes := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "parse warning: %s\n", e)
		}
	}

	vrbPath := strings.TrimSuffix(wrlPath, filepath.Ext(wrlPath)) + ".vrb"
	out, err := os.Create(vrbPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close error: %v\n", err)
		}
	}()

	if err := serializer.Encode(out, nodes); err != nil {
		_ = os.Remove(vrbPath)
		return err
	}

	info, _ := out.Stat()
	fmt.Printf("%s -> %s (%d bytes)\n", wrlPath, vrbPath, info.Size())
	return nil
}

func decodeFile(vrbPath string) error {
	f, err := os.Open(vrbPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close error: %v\n", err)
		}
	}()

	nodes, err := serializer.Decode(f)
	if err != nil {
		return err
	}

	fmt.Printf("decoded %d top-level nodes from %s\n", len(nodes), vrbPath)
	return nil
}
