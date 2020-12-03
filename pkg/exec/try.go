// +build !js

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

package exec

import (
	"io"
	"path/filepath"

	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/wellknownerrors"
)

var (
	tryCommands = make(map[string]tryFunc)
)

// DoIfTryFailed will first try to handle command internally, if the command is not handled or failed to handle,
// execute it directly on host
func DoIfTryFailed(
	stdin io.Reader,
	stdout, stderr io.Writer,
	command []string,
	tty bool,
	env map[string]string,
	tryOnly bool,
) (Cmd, error) {
	var (
		err error
		cmd Cmd
	)

	bin := filepath.Base(command[0])
	tryExec, ok := tryCommands[bin]
	if ok {
		// can try this command, do it
		cmd, err = tryExec(stdin, stdout, stderr, command, tty)
		if err == nil {
			// handled
			return cmd, nil
		}

		if tryOnly {
			return nil, err
		}
	}

	if tryOnly {
		return nil, wellknownerrors.ErrNotSupported
	}

	// not handled, do it locally
	return exechelper.Do(exechelper.Spec{
		Context: nil,
		Env:     env,
		Command: command,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		Tty:     tty,
	})
}
