package clang

import (
	"sync"

	ts "github.com/wow-look-at-my/go-tree-sitter"
)

// load is set by the parser.gen.go the generate step writes. The tables are
// data, generated at build time, so this half compiles without them and a
// consumer resolving this module from the proxy still gets a package.
var (
	load      func() *ts.Language
	loadOnce  sync.Once
	generated *ts.Language
)

// Language returns the grammar, decoding its tables when a parse needs them.
// It panics when the generate step has not run, rather than hand back a
// language that parses nothing.
func Language() *ts.Language {
	loadOnce.Do(func() {
		if load != nil {
			generated = load()
		}
	})
	if generated == nil {
		panic("clang: parser.gen.go is missing. Run: go generate ./grammars/clang")
	}
	return generated
}
