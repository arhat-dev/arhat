// +build !windows,!js,!plan9,!aix

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
	"io"
	"os/exec"

	"github.com/creack/pty"
)

func startCmdWithTty(
	cmd *exec.Cmd,
	stdin io.Reader,
	stdout io.Writer,
	onCopyErr func(error),
) (resizeFunc, func(), error) {
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, nil, err
	}

	if stdin != nil {
		go func() {
			_, err := io.Copy(f, stdin)
			if err != nil {
				onCopyErr(err)
			}
		}()
	}

	if stdout != nil {
		go func() {
			_, err := io.Copy(stdout, f)
			if err != nil {
				onCopyErr(err)
			}
		}()
	}

	return func(cols, rows uint32) error {
		return pty.Setsize(f, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	}, func() { _ = f.Close() }, nil
}
