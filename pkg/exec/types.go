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

// tryFunc for all kinds of command try handler
type tryFunc func(stdin io.Reader, stdout, stderr io.Writer, command []string, tty bool) (Cmd, error)

type Cmd interface {
	Resize(cols, rows uint32) error
	Wait() (int, error)
}

type flexCmd struct {
	do func() error
}

func (c *flexCmd) Resize(cols, rows uint32) error {
	return nil
}

func (c *flexCmd) Wait() (int, error) {
	err := c.do()
	if err != nil {
		return exechelper.DefaultExitCodeOnError, err
	}

	return 0, nil
}
