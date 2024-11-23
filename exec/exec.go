package exec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Command(name string, args ...string) (string, error) {
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running cmd: %s %s, output: %s, err: %v",
			name, strings.Join(args, " "), strings.TrimSpace(string(output)), err)
	}
	return strings.TrimSpace(string(output)), nil
}

func InteractiveCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // Allow user to interact with the process

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute interactive cmd: %v", err)
	}

	return nil
}
