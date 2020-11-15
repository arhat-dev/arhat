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

	"arhat.dev/pkg/exechelper"
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
) (*exechelper.Cmd, error) {
	var err error
	tryExec, ok := tryCommands[command[0]]
	if ok {
		err = tryExec(stdin, stdout, stderr, command, tty)
	}

	if ok && err == nil {
		// handled
		return nil, nil
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
