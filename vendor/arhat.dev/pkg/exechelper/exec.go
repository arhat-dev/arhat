package exechelper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"
)

type Spec struct {
	Context context.Context
	Logger  log.Interface

	Env     map[string]string
	Command []string

	Stdin          io.Reader
	Stdout, Stderr io.Writer

	Tty            bool
	OnResizeSignal TtyResizeSignalFunc
}

const (
	DefaultExitCodeOnError = 128
)

type TtyResizeSignalFunc func(doResize func(cols, rows uint64) error) (more bool)

func DoHeadless(command []string, env map[string]string) (int, error) {
	return Do(Spec{
		Env:     env,
		Command: command,
		Tty:     false,
	})
}

func Prepare(ctx context.Context, command []string, tty bool, env map[string]string) *exec.Cmd {
	var cmd *exec.Cmd
	if ctx == nil {
		cmd = exec.Command(command[0], command[1:]...)
	} else {
		cmd = exec.CommandContext(ctx, command[0], command[1:]...)
	}

	// if using tty in unix, github.com/creack/pty will Setsid, and if we
	// Setpgid, will fail the process creation
	cmd.SysProcAttr = getSysProcAttr(tty)

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	return cmd
}

// Do execute command directly in host
func Do(s Spec) (int, error) {
	if len(s.Command) == 0 {
		// impossible for agent exec, but still check for test
		return DefaultExitCodeOnError, fmt.Errorf("empty command: %w", wellknownerrors.ErrInvalidOperation)
	}

	cmd := Prepare(s.Context, s.Command, s.Tty, s.Env)
	if s.Tty {
		cleanup, err := startCmdWithTty(s.Logger, cmd, s.Stdin, s.Stdout, s.OnResizeSignal)
		if err != nil {
			return DefaultExitCodeOnError, err
		}
		defer cleanup()
	} else {
		cmd.Stdin = s.Stdin
		cmd.Stdout = s.Stdout
		cmd.Stderr = s.Stderr

		if err := cmd.Start(); err != nil {
			return DefaultExitCodeOnError, err
		}
	}

	if err := cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), err
		}

		if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrClosedPipe) {
			return DefaultExitCodeOnError, err
		}
	}

	return 0, nil
}
