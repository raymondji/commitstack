package exec

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Output struct {
	ExitCode  int
	ExitError *exec.ExitError
	Stdout    string
	Stderr    string
}

func (o Output) Lines() []string {
	s := o.Stdout
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

type runOpts struct {
	Args []string
	Env  []string
	Dir  string
	// If true, don't return an error if the command exited with a non-zero
	// exit code.
	IgnoreExitError bool
	// If true, the standard I/Os are connected to the console, allowing the git command to
	// interact with the user. Stdout and Stderr will be empty.
	Interactive bool
}

type runOpt func(*runOpts)

func WithArgs(args ...string) runOpt {
	return func(opts *runOpts) {
		opts.Args = args
	}
}

func WithEnv(vars ...string) runOpt {
	return func(opts *runOpts) {
		opts.Env = vars
	}
}

func WithInteractive(interactive bool) runOpt {
	return func(opts *runOpts) {
		opts.Interactive = interactive
	}
}

func WithDir(dir string) runOpt {
	return func(opts *runOpts) {
		opts.Dir = dir
	}
}

func Run(name string, fOpts ...runOpt) (*Output, error) {
	var opts runOpts
	for _, o := range fOpts {
		o(&opts)
	}

	cmd := exec.Command(name, opts.Args...)
	cmd.Dir = opts.Dir
	var stdout, stderr bytes.Buffer
	if opts.Interactive {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}
	cmd.Env = append(os.Environ(), opts.Env...)
	err := cmd.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return nil, fmt.Errorf("%s %s, err: %v", name, opts.Args, err)
	}
	if err != nil && !opts.IgnoreExitError && exitError.ExitCode() != 0 {
		// ExitError.Stderr is only populated if the command was started without
		// a Stderr pipe, which is not the case here. Just populate it ourselves
		// to make it easier for callers to access.
		exitError.Stderr = stderr.Bytes()
		return nil, fmt.Errorf("%s %s (%s), err: %v", name, opts.Args, stderr.String(), err)
	}
	return &Output{
		ExitCode:  cmd.ProcessState.ExitCode(),
		ExitError: exitError,
		Stdout:    strings.TrimSpace(stdout.String()),
		Stderr:    strings.TrimSpace(stderr.String()),
	}, nil
}

func InPath(name string) (bool, error) {
	_, err := exec.LookPath(name)
	if err != nil {
		return false, errors.New("not installed or not in PATH")
	}
	return true, nil
}
