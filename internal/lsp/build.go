package lsp

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gnolang/gnopls/internal/tools"
)

type ErrorInfo struct {
	FileName string
	Line     int
	Column   int
	Span     []int
	Msg      string
	Tool     string
}

func (s *server) Transpile() ([]ErrorInfo, error) {
	moduleName := filepath.Base(s.workspaceFolder)
	tmpDir := filepath.Join(s.env.GNOHOME, "gnopls", "tmp", moduleName)

	err := copyDir(s.workspaceFolder, tmpDir)
	if err != nil {
		return nil, err
	}

	preOut, errTranspile := tools.Transpile(tmpDir)
	if len(preOut) > 0 {
		// parse errors even if errTranspile!=nil bc that's always the case if
		// there's errors to parse.
		slog.Info("transpile error", "out", string(preOut), "err", errTranspile)
		errors, errParse := s.parseErrors(string(preOut), "transpile")
		if errParse != nil {
			return nil, errParse
		}
		if len(errors) == 0 && errTranspile != nil {
			// no parsed errors but errTranspile!=nil, this is an unexpected error.
			// (for example the gno binary was not found)
			return nil, errTranspile
		}
		return errors, nil
	}
	return nil, nil
}

// This is used to extract information from the `gno build` command
// (see `parseError` below).
//
// TODO: Maybe there's a way to get this in a structured format?
// -> will be available in go1.24 with `go build -json`
var errorRe = regexp.MustCompile(`(?m)^([^#]+?):(\d+):(\d+):(.+)$`)

// parseErrors parses the output of the `gno transpile -gobuild` command for
// errors.
//
// The format is:
// ```
// <file.gno>:<line>:<col>: <error>
// ```
func (s *server) parseErrors(output, cmd string) ([]ErrorInfo, error) {
	errors := []ErrorInfo{}

	matches := errorRe.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return errors, nil
	}

	for _, match := range matches {
		path := match[1]
		line, err := strconv.Atoi(match[2])
		if err != nil {
			return nil, fmt.Errorf("parseErrors '%s': %w", match, err)
		}

		column, err := strconv.Atoi(match[3])
		if err != nil {
			return nil, fmt.Errorf("parseErrors '%s': %w", match, err)
		}
		msg := strings.TrimSpace(match[4])
		slog.Debug("parseErrors", "path", path, "line", line, "column", column, "msg", msg)
		errors = append(errors, ErrorInfo{
			FileName: filepath.Join(s.workspaceFolder, path),
			Line:     line,
			Column:   column,
			Span:     []int{column, column},
			Msg:      msg,
			Tool:     cmd,
		})
	}

	return errors, nil
}

// NOTE(tb): This function tries to guess the column start and end of the
// error, by splitting the line into tokens (space separated). While this
// might work most of the time, in some cases it might not, so I think
// it is preferable not to try to guess and just return:
// column_start = column_end
// just like the output of gno transpile which only returns the start of the
// column. This is why this function is no longer invoked in parseError.
//
// findError finds the error in the document, shifting the line and column
// numbers to account for the header information in the generated Go file.
func findError(file *GnoFile, fname string, line, col int, msg string, tool string) ErrorInfo {
	msg = strings.TrimSpace(msg)
	// TODO: can be removed?
	// see: https://github.com/gnolang/gno/pull/1670
	if tool == "transpile" {
		// fname parsed from transpile result can be incorrect
		// e.g filename = `filename.gno: transpile: parse: tmp.gno`
		parts := strings.Split(fname, ":")
		fname = parts[0]
	}

	// Error messages are of the form:
	//
	// <token> <error> (<info>)
	// <error>: <token>
	//
	// We want to strip the parens and find the token in the file.
	parens := regexp.MustCompile(`\((.+)\)`)
	needle := parens.ReplaceAllString(msg, "")
	tokens := strings.Fields(needle)

	errorInfo := ErrorInfo{
		FileName: strings.TrimPrefix(GoToGnoFileName(filepath.Base(fname)), "."),
		Line:     line,
		Column:   col,
		Span:     []int{0, 0},
		Msg:      msg,
		Tool:     tool,
	}

	lines := strings.SplitAfter(string(file.Src), "\n")
	for i, l := range lines {
		if i != line-1 { // zero-indexed
			continue
		}
		for _, token := range tokens {
			tokRe := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(token)))
			if tokRe.MatchString(l) {
				errorInfo.Line = i + 1
				errorInfo.Span = []int{col, col + len(token)}
				return errorInfo
			}
		}
	}

	// If we couldn't find the token, just return the original error + the
	// full line.
	errorInfo.Span = []int{col, col + 1}

	return errorInfo
}
