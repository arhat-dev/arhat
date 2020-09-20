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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"arhat.dev/pkg/exechelper"
)

const (
	binUmount     = "umount"
	binFusermount = "fusermount"
)

func Lookup(bin string, extraLookupPaths []string) (string, error) {
	if len(extraLookupPaths) != 0 {
		for _, lookupPath := range extraLookupPaths {
			binPath := filepath.Join(lookupPath, bin)
			info, err := os.Stat(binPath)
			if err != nil {
				return "", fmt.Errorf("unable to check %s in path %s: %w", bin, lookupPath, err)
			}

			if m := info.Mode(); !m.IsDir() && m.Perm()&0111 != 0 {
				return "", fmt.Errorf("%s in path %s is not a executable file", bin, lookupPath)
			}

			// nolint:staticcheck
			return binPath, nil
		}
	}

	binPath, err := exec.LookPath(bin)
	if err != nil {
		return "", fmt.Errorf("unable to find executable %s: %w", bin, err)
	}

	return binPath, nil
}

func LookupUnmountUtil(extraLookupPaths []string, fuse bool) (string, error) {
	var bin string
	switch runtime.GOOS {
	case "linux":
		if fuse {
			bin = binFusermount
		} else {
			bin = binUmount
		}
	default:
		bin = binUmount
	}

	return Lookup(bin, extraLookupPaths)
}

func Unmount(binPath string, mountPoint string) error {
	command := []string{binPath}
	switch filepath.Base(binPath) {
	case binFusermount:
		// linux only
		command = append(command, "-z", "-u", mountPoint)
	case binUmount:
		switch runtime.GOOS {
		case "linux":
			// lazy unmount is not supported on darwin
			command = append(command, "-l")
		default:
			command = append(command, "-f")
		}
		command = append(command, mountPoint)
	default:
		return fmt.Errorf("unknown fuse unmounter: %s", binPath)
	}

	_, err := exechelper.DoHeadless(command, nil)
	return err
}

func prepareFile(def *os.File, f string) (io.Writer, error) {
	switch strings.ToLower(f) {
	case "":
		return def, nil
	case "none":
		return ioutil.Discard, nil
	case "stderr":
		return os.Stderr, nil
	case "stdout":
		return os.Stdout, nil
	default:
		file, err := os.OpenFile(f, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
}

func Execute(
	command []string,
	stdoutFile, stderrFile string,
	wait time.Duration,
	onExited func(err error),
) (*os.Process, error) {
	cmd := exechelper.Prepare(context.TODO(), command, false, nil)

	var err error
	cmd.Stdout, err = prepareFile(os.Stdout, stdoutFile)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stdout file: %w", err)
	}

	cmd.Stderr, err = prepareFile(os.Stderr, stderrFile)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stderr file: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to execute sshfs: %w", err)
	}

	errCh := make(chan error)
	go func() {
		timer := time.NewTimer(wait)
		defer timer.Stop()

		go func() {
			err := cmd.Wait()

			defer func() {
				// recover from possible send on closed chan panic
				_ = recover()
				onExited(err)
			}()

			errCh <- err
		}()

		// nolint:gosimple
		select {
		case <-timer.C:
			close(errCh)
		}
	}()

	return cmd.Process, <-errCh
}
