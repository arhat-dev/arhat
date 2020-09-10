// +build windows

/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
	"os"
)

// IsLikelyNotMountPoint determines if a directory is not a mountpoint.
func IsLikelyNotMountPoint(file string) (bool, error) {
	stat, err := os.Lstat(file)
	if err != nil {
		return true, err
	}
	// If current file is a symlink, then it is a mountpoint.
	if stat.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(file)
		if err != nil {
			return true, fmt.Errorf("readlink error: %w", err)
		}
		exi, err := exists(target)
		if err != nil {
			return true, err
		}
		return !exi, nil
	}

	return true, nil
}

func exists(filename string) (bool, error) {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
