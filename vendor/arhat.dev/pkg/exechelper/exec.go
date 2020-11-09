/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package exechelper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

type Spec struct {
	Context context.Context

	Env     map[string]string
	Command []string

	ExtraLookupPaths []string

	Stdin          io.Reader
	Stdout, Stderr io.Writer

	Tty            bool
	OnTtyCopyError func(err error)
}

const (
	DefaultExitCodeOnError = 128
)

type resizeFunc func(cols, rows int64) error

type Cmd struct {
	ExecCmd *exec.Cmd

	doResize resizeFunc
	cleanup  func()
}

// Resize tty windows if was created with tty enabled
func (c *Cmd) Resize(cols, rows int64) error {
	if c.doResize != nil {
		return c.doResize(cols, rows)
	}

	return nil
}

// Release process if wait was not called and you want to terminate it
func (c *Cmd) Release() error {
	if c.ExecCmd.Process != nil {
		return c.ExecCmd.Process.Release()
	}

	return nil
}

// Wait until command exited
func (c *Cmd) Wait() (int, error) {
	err := c.ExecCmd.Wait()

	if c.cleanup != nil {
		c.cleanup()
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), exitError
		}

		// could not get exit code
		if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrClosedPipe) {
			return DefaultExitCodeOnError, err
		}
	}

	return 0, nil
}

func DoHeadless(command []string, env map[string]string) (int, error) {
	cmd, err := Do(Spec{
		Env:     env,
		Command: command,
		Tty:     false,
	})
	if err != nil {
		return DefaultExitCodeOnError, err
	}

	return cmd.Wait()
}

// Prepare an unstarted exec.Cmd
func Prepare(
	ctx context.Context,
	command, extraPaths []string,
	tty bool,
	env map[string]string,
) (*exec.Cmd, error) {
	if len(command) == 0 {
		// defensive check
		return nil, fmt.Errorf("empty command")
	}

	bin, err := Lookup(command[0], extraPaths)
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if ctx == nil {
		cmd = exec.Command(bin, command[1:]...)
	} else {
		cmd = exec.CommandContext(ctx, bin, command[1:]...)
	}

	// if using tty in unix, github.com/creack/pty will Setsid, and if we
	// Setpgid, will fail the process creation
	cmd.SysProcAttr = getSysProcAttr(tty)

	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	return cmd, nil
}

// Do execute command directly in host
func Do(s Spec) (*Cmd, error) {
	cmd, err := Prepare(s.Context, s.Command, s.ExtraLookupPaths, s.Tty, s.Env)
	if err != nil {
		return nil, err
	}

	startedCmd := &Cmd{
		ExecCmd: cmd,
	}

	if s.Tty {
		handErr := s.OnTtyCopyError
		if handErr == nil {
			handErr = func(error) {}
		}

		startedCmd.doResize, startedCmd.cleanup, err = startCmdWithTty(cmd, s.Stdin, s.Stdout, handErr)
		if err != nil {
			if cmd.Process != nil {
				_ = cmd.Process.Release()
			}

			return nil, err
		}

		return startedCmd, nil
	}

	cmd.Stdin = s.Stdin
	cmd.Stdout = s.Stdout
	cmd.Stderr = s.Stderr

	if err := cmd.Start(); err != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Release()
		}

		return nil, err
	}

	return startedCmd, nil
}
