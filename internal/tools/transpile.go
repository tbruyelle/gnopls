package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Transpile a Gno package: gno transpile <dir>.
func Transpile(rootDir string) ([]byte, error) {
	cmd := exec.Command("gno", "transpile", "-skip-imports", "-gobuild", filepath.Join(rootDir))
	// FIXME(tb): See https://github.com/gnolang/gno/pull/1695/files#r1697255524
	const disableGoMod = "GO111MODULE=off"
	cmd.Env = append(os.Environ(), disableGoMod)
	bz, err := cmd.CombinedOutput()
	if err != nil {
		return bz, fmt.Errorf("running '%s': %w: %s", strings.Join(cmd.Args, " "), err, string(bz))
	}
	return bz, nil
}
