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

package storageutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/iohelper"
	"arhat.dev/pkg/queue"
)

var (
	ErrMountpointInUse     = errors.New("already in use")
	ErrMountpointInProcess = errors.New("already in process")
)

func NewClient(
	ctx context.Context,
	impl Interface,
	successTimeWait time.Duration,
	extraLookupPaths []string,
	stdoutFile, stderrFile string,
) (_ *Client, err error) {
	stdout, err := prepareFile(os.Stdout, stdoutFile)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stdout file: %w", err)
	}

	defer func() {
		if err != nil {
			stdout.Close()
		}
	}()

	stderr, err := prepareFile(os.Stderr, stderrFile)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stderr file: %w", err)
	}

	tq := queue.NewTimeoutQueue()
	tq.Start(ctx.Done())

	cli := &Client{
		impl: impl,

		successTimeWait:  successTimeWait,
		extraLookupPaths: extraLookupPaths,

		stdout: stdout,
		stderr: stderr,

		mounted: make(map[string]*exechelper.Cmd),

		tq: tq,

		mu: new(sync.RWMutex),
	}

	go cli.routine()

	return cli, nil
}

type Client struct {
	impl Interface

	successTimeWait  time.Duration
	extraLookupPaths []string

	stdout io.WriteCloser
	stderr io.WriteCloser

	mounted map[string]*exechelper.Cmd

	tq *queue.TimeoutQueue

	mu *sync.RWMutex
}

func (c *Client) routine() {
	ch := c.tq.TakeCh()
	for t := range ch {
		errCh, ok := t.Data.(chan error)
		if !ok {
			continue
		}

		func() {
			defer func() {
				// defensive, should be panic in any case
				_ = recover()
			}()

			close(errCh)
		}()
	}
}

func (c *Client) Mount(
	ctx context.Context,
	remotePath, mountPoint string,
	onExited ExitHandleFunc,
) error {
	c.mu.Lock()
	_, mounted := c.mounted[mountPoint]
	if mounted {
		c.mu.Unlock()
		return ErrMountpointInUse
	}

	_, mounting := c.tq.Find(mountPoint)
	if mounting {
		c.mu.Unlock()
		return ErrMountpointInProcess
	}

	cmd := c.impl.GetMountCmd(remotePath, mountPoint)
	if len(cmd) == 0 {
		return fmt.Errorf("invalid empty mount command")
	}

	startedCmd, err := exechelper.Do(exechelper.Spec{
		Command: cmd,

		Stdout: c.stdout,
		Stderr: c.stderr,

		ExtraLookupPaths: c.extraLookupPaths,
	})
	if err != nil {
		c.mu.Unlock()
		return err
	}

	errCh := make(chan error)
	// close errCh after successTimeWait
	_ = c.tq.OfferWithDelay(mountPoint, errCh, c.successTimeWait)

	// mark it as mounted
	c.mounted[mountPoint] = startedCmd
	c.mu.Unlock()

	// start mount success check, it's time consuming, do not hold the lock
	// start a goroutine to monitor process and notify when process exited
	go func() {
		// wait until command exited
		_, err := startedCmd.Wait()

		defer func() {
			// recover from possible send on closed chan panic
			chClosed := recover()

			// command exited, mark it as not mounted
			c.mu.Lock()
			delete(c.mounted, mountPoint)
			c.mu.Unlock()

			if chClosed != nil {
				// successful return, unexpected exit
				onExited(remotePath, mountPoint, err)
			}
		}()

		select {
		case <-ctx.Done():
			return
		case errCh <- err:
			// this can panic if errCh was closed in routine
		}
	}()

	select {
	case <-ctx.Done():
		_ = startedCmd.ExecCmd.Process.Kill()
		return ctx.Err()
	case err := <-errCh:
		// err can only happen when command exited with error
		// otherwise we have passed initial success wait
		if err != nil {
			c.mu.Lock()
			delete(c.mounted, mountPoint)
			c.mu.Unlock()
			return err
		}
	}

	return nil
}

func (c *Client) Unmount(ctx context.Context, mountPoint string) error {
	c.mu.Lock()

	startedCmd, ok := c.mounted[mountPoint]
	if !ok {
		c.mu.Unlock()
		return nil
	}

	delete(c.mounted, mountPoint)
	c.mu.Unlock()

	_ = startedCmd.Release()

	cmd := c.impl.GetUnmountCmd(mountPoint)
	if len(cmd) == 0 {
		return fmt.Errorf("invalid empty unmount command")
	}

	_, err := exechelper.DoHeadless(cmd, nil)
	return err
}

func (c *Client) Close() {
	_ = c.stdout.Close()
	_ = c.stderr.Close()
}

func prepareFile(def *os.File, f string) (io.WriteCloser, error) {
	switch strings.ToLower(f) {
	case "":
		return def, nil
	case "none":
		return iohelper.NopWriteCloser(ioutil.Discard), nil
	case "stderr":
		return iohelper.NopWriteCloser(os.Stderr), nil
	case "stdout":
		return iohelper.NopWriteCloser(os.Stdout), nil
	default:
		file, err := os.OpenFile(f, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
}
