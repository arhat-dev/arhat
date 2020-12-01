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
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

type Spec struct {
	Context context.Context

	Env         map[string]string
	Command     []string
	SysProcAttr *syscall.SysProcAttr

	ExtraLookupPaths []string

	// stdin/stdout/stderr streams, not used if tty required
	Stdin          io.Reader
	Stdout, Stderr io.Writer

	Tty bool
}

const (
	DefaultExitCodeOnError = 128
)

type resizeFunc func(cols, rows uint32) error

type Cmd struct {
	ExecCmd *exec.Cmd

	// stdin/stdout set when created with tty enabled
	TtyInput  io.WriteCloser
	TtyOutput io.ReadCloser

	doResize resizeFunc
	cleanup  func()
}

// Resize tty windows if was created with tty enabled
func (c *Cmd) Resize(cols, rows uint32) error {
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
		return DefaultExitCodeOnError, err
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
		// unable to start command
		return DefaultExitCodeOnError, err
	}

	return cmd.Wait()
}

// Prepare an unstarted exec.Cmd
func Prepare(s Spec) (*exec.Cmd, error) {
	if len(s.Command) == 0 {
		// defensive check
		return nil, fmt.Errorf("empty command")
	}

	bin, err := Lookup(s.Command[0], s.ExtraLookupPaths)
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if s.Context == nil {
		cmd = exec.Command(bin, s.Command[1:]...)
	} else {
		cmd = exec.CommandContext(s.Context, bin, s.Command[1:]...)
	}

	cmd.SysProcAttr = getSysProcAttr(s.Tty, s.SysProcAttr)

	for k, v := range s.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	return cmd, nil
}

// Do execute command directly in host
func Do(s Spec) (*Cmd, error) {
	cmd, err := Prepare(s)
	if err != nil {
		return nil, err
	}

	startedCmd := &Cmd{
		ExecCmd: cmd,
	}

	if s.Tty {
		startedCmd.doResize,
			startedCmd.cleanup,
			startedCmd.TtyInput,
			startedCmd.TtyOutput, err = startCmdWithTty(cmd)
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
