package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Exec executes a git command with the given arguments.
// It returns the combined stdout and stderr output, and an error if the command fails.
func Exec(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git command failed: %s %s: %w", strings.Join(args, " "), stderr.String(), err)
	}

	return stdout.String(), nil
}
