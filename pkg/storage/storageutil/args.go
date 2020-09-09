/*
Copyright 2019 The arhat.dev Authors.

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
	"os"
	"path/filepath"

	"arhat.dev/arhat/pkg/constant"
)

func ResolveCommand(bin string, args []string, remotePath, localPath string) ([]string, error) {
	command := []string{bin}

	remotePath = filepath.Clean(remotePath)

	envMapping := map[string]string{
		constant.StorageArgEnvLocalPath:  localPath,
		constant.StorageArgEnvRemotePath: remotePath,
	}

	for _, a := range args {
		command = append(command, os.Expand(a, func(s string) string {
			return envMapping[s]
		}))
	}

	return command, nil
}
