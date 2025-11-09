package util

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"time"
)

// RunCommand runs a shell command with a timeout and returns output, exit code, and error
func RunCommand(ctx context.Context, command string, timeout time.Duration) (string, int, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// ✅ run through Windows' cmd.exe
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		// ✅ run through bash on Unix-like systems
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return out.String(), exitCode, err
}
