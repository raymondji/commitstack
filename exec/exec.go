package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Command(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running cmd: %s %s, stderr: %s, err: %v",
			name, strings.Join(args, " "), stderr.String(), err)
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
