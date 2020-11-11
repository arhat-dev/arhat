// +build !noflaghelper

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
	"time"

	"github.com/spf13/pflag"
)

func FlagsForClient(prefix string, config *ClientConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("storage.client", pflag.ExitOnError)

	fs.StringVar(&config.Driver, prefix+"driver",
		"", "set storage driver to use",
	)

	fs.StringVar(&config.StdoutFile, prefix+"stdoutFile", "stdout", "set command stdout file")
	fs.StringVar(&config.StderrFile, prefix+"stderrFile", "stderr", "set command stderr file")

	fs.DurationVar(&config.SuccessTimeWait, prefix+"successTimeWait",
		5*time.Second, "set time to wait before treat cmd exec as successful",
	)

	fs.StringSliceVar(&config.ExtraLookupPaths, prefix+"extraLookupPaths",
		[]string{}, "set extra paths for binary lookup",
	)

	return fs
}
