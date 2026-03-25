package parser

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// parseURLList parses the URL list after an EXTERNPROTO interface.
// Formats: "url.wrl" or [ "url1.wrl" "url2.wrl" ]
func (p *Parser) parseURLList() []string {
	var urls []string
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokString {
				p.lex.Next()
				urls = append(urls, p.lex.StrVal())
			} else {
				p.lex.Next()
			}
		}
		if p.lex.Peek() == TokCloseBracket {
			p.lex.Next()
		}
	} else if p.lex.Peek() == TokString {
		p.lex.Next()
		urls = append(urls, p.lex.StrVal())
	}
	return urls
}

// resolveExternProto fetches the first available URL and fills in
// the PROTO body for an EXTERNPROTO stub. Returns true if resolved.
func (p *Parser) resolveExternProto(def *ProtoDefinition) bool {
	for _, rawURL := range def.URLs {
		var fileName, protoName string
		if idx := strings.LastIndex(rawURL, "#"); idx >= 0 {
			fileName = rawURL[:idx]
			protoName = rawURL[idx+1:]
		} else {
			fileName = rawURL
			protoName = def.Name
		}

		if fileName == "" || protoName == "" {
			continue
		}

		r, err := p.fetchURL(fileName)
		if err != nil {
			continue
		}

		found := p.findProtoInReader(r, protoName)
		r.Close()
		if found != nil {
			def.Body = found.Body
			def.Fields = found.Fields
			return true
		}
	}
	return false
}

// fetchURL opens a URL, using the injected fetcher or the default.
func (p *Parser) fetchURL(rawURL string) (io.ReadCloser, error) {
	if p.urlFetcher != nil {
		return p.urlFetcher(rawURL)
	}
	return defaultFetchURL(p.baseDir, rawURL)
}

// defaultFetchURL handles local files and HTTP URLs.
func defaultFetchURL(baseDir, rawURL string) (io.ReadCloser, error) {
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(rawURL)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, rawURL)
		}
		return resp.Body, nil
	}

	path := rawURL
	if !filepath.IsAbs(path) && baseDir != "" {
		path = filepath.Join(baseDir, path)
	}
	return os.Open(path)
}

// findProtoInReader parses a VRML source and returns the named PROTO, or nil.
func (p *Parser) findProtoInReader(r io.Reader, protoName string) *ProtoDefinition {
	sub := NewParser(r)
	sub.baseDir = p.baseDir
	sub.urlFetcher = p.urlFetcher
	sub.Parse()
	// Propagate all proto definitions from the resolved file
	for k, v := range sub.protoTable {
		if _, exists := p.protoTable[k]; !exists {
			p.protoTable[k] = v
		}
	}
	if def, ok := sub.protoTable[protoName]; ok {
		return def
	}
	return nil
}
