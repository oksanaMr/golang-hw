package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	for key, value := range env {
		if value.NeedRemove {
			os.Unsetenv(key)
		} else {
			os.Unsetenv(key)
			os.Setenv(key, value.Value)
		}
	}

	//nolint:gosec
	newCmd := exec.Command(cmd[0], cmd[1:]...)
	newCmd.Stdin = os.Stdin
	newCmd.Stdout = os.Stdout
	newCmd.Stderr = os.Stderr
	newCmd.Env = os.Environ()

	if err := newCmd.Run(); err != nil {
		fmt.Print(err)
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
		return 1
	}
	return 0
}
