// +build !noexectry,!noexectry_test
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
	"fmt"
	"io"
	"os"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"
	"github.com/spf13/pflag"
)

func init() {
	tryCommands["test"] = tryTestCmd
}

// tryTestCmd handle test command execution issued by `kubectl cp`
//
// returned error will be wellknownerrors.ErrNotSupported if the test command is not
// checking whether the path is a directory
func tryTestCmd(
	_ io.Reader,
	_, _ io.Writer,
	_ <-chan *aranyagopb.TerminalResizeCmd,
	command []string,
	_ bool,
) error {
	if len(command) != 3 {
		return wellknownerrors.ErrNotSupported
	}

	var (
		testingDir bool
	)

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.BoolVarP(&testingDir, "dir", "d", false, "")

	if err := flags.Parse(command[1:]); err != nil {
		return wellknownerrors.ErrNotSupported
	}

	if !testingDir {
		return wellknownerrors.ErrNotSupported
	}

	dirName := command[2]
	f, err := os.Stat(dirName)
	if err != nil {
		return err
	}

	if !f.IsDir() {
		return fmt.Errorf("path is not a dir")
	}

	return nil
}
