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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Lookup(bin string, extraLookupPaths []string) (string, error) {
	if strings.Contains(bin, "/") {
		return exec.LookPath(bin)
	}

	if filepath.Base(bin) == bin {
		for _, lookupPath := range extraLookupPaths {
			binPath := filepath.Join(lookupPath, bin)
			info, err := os.Stat(binPath)
			if err != nil {
				// unable to check file status
				continue
			}

			// found file in this path, check if valid
			m := info.Mode()
			switch {
			case m.IsDir():
			case m.Perm()&0111 == 0:
			default:
				return binPath, nil
			}
		}
	}

	return exec.LookPath(bin)
}
