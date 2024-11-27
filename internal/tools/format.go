package tools

import (
	"os/exec"
)

func Format(file string) ([]byte, error) {
	return exec.Command("gno", "fmt", file).CombinedOutput()
}
