package lsp

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"log/slog"
	"strings"
	"unicode/utf8"

	"go.lsp.dev/protocol"
	"golang.org/x/mod/modfile"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type Snapshot struct {
	file cmap.ConcurrentMap[string, *GnoFile]
}

func NewSnapshot() *Snapshot {
	return &Snapshot{
		file: cmap.New[*GnoFile](),
	}
}

func (s *Snapshot) Get(filePath string) (*GnoFile, bool) {
	return s.file.Get(filePath)
}

// contains gno file.
type GnoFile struct {
	URI protocol.DocumentURI
	Src []byte
}

// contains parsed gno file.
type ParsedGnoFile struct {
	URI    protocol.DocumentURI
	File   *ast.File
	Fset   *token.FileSet
	Errors scanner.ErrorList

	Src []byte
}

// ParseGno reads src from disk and parse it
func (f *GnoFile) ParseGno(ctx context.Context) *ParsedGnoFile {
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, f.URI.Filename(), nil, parser.ParseComments)
	var parseErr scanner.ErrorList
	if err != nil {
		parseErr = err.(scanner.ErrorList) //nolint:errcheck,errorlint
	}

	return &ParsedGnoFile{
		URI:    f.URI,
		File:   ast,
		Fset:   fset,
		Errors: parseErr,
		Src:    f.Src,
	}
}

// ParseGno2 parses src from GnoFile instead of reading from disk
// Right now it's only used in `completion.go`
// TODO: Replace content of `ParseGno` with `ParseGno2`
func (f *GnoFile) ParseGno2(ctx context.Context) *ParsedGnoFile {
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, f.URI.Filename(), f.Src, parser.ParseComments)
	var parseErr scanner.ErrorList
	if err != nil {
		parseErr = err.(scanner.ErrorList) //nolint:errcheck,errorlint
	}
	return &ParsedGnoFile{
		URI:    f.URI,
		File:   ast,
		Fset:   fset,
		Errors: parseErr,
		Src:    f.Src,
	}
}

// contains parsed gno.mod file.
type ParsedGnoMod struct {
	URI  string
	File *modfile.File
}

func (f *GnoFile) TokenAt(pos protocol.Position) (*HoveredToken, error) {
	lines := strings.SplitAfter(string(f.Src), "\n")

	size := uint32(len(lines))
	if pos.Line >= size {
		return nil, errors.New("line out of range")
	}

	line := lines[pos.Line]
	lineLen := uint32(len(line))

	// TODO: fix it. should not happen?
	if len(line) == 0 {
		return nil, errors.New("no token found")
	}

	index := pos.Character
	start := index
	// TODO: fix it. should not happen?
	if lineLen < start {
		return nil, errors.New("start is greater than len")
	}
	for start > 0 && line[start-1] != ' ' {
		start--
	}

	end := index
	slog.Info(fmt.Sprintf("curser at: %d", end))
	for end < lineLen && line[end] != ' ' {
		end++
	}

	if start == end {
		return nil, errors.New("no token found")
	}

	return &HoveredToken{
		Text:  line[start:end],
		Start: int(start),
		End:   int(end),
	}, nil
}

func (f *GnoFile) PositionToOffset(pos protocol.Position) int {
	lines := strings.SplitAfter(string(f.Src), "\n")
	offset := 0
	for i, l := range lines {
		if i == int(pos.Line) {
			break
		}
		offset += utf8.RuneCountInString(l)
	}
	return offset + int(pos.Character)
}
