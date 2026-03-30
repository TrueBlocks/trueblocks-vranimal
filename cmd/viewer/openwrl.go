package main

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// openWRL opens a .wrl file for reading. If the path ends with .gz or .wrz,
// the returned reader transparently decompresses gzip data.
// The caller must call the returned close function when done.
func openWRL(path string) (io.Reader, func() error, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, nil, err
	}

	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".gz") || strings.HasSuffix(lower, ".wrz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			_ = f.Close()
			return nil, nil, err
		}
		closer := func() error {
			e1 := gz.Close()
			e2 := f.Close()
			if e1 != nil {
				return e1
			}
			return e2
		}
		return gz, closer, nil
	}

	return f, f.Close, nil
}
