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
	"path/filepath"
	"runtime"

	"arhat.dev/pkg/envhelper"
	"arhat.dev/pkg/exechelper"
)

const (
	StorageArgEnvRemotePath = "ARHAT_STORAGE_REMOTE_PATH"
	StorageArgEnvMountpoint = "ARHAT_STORAGE_MOUNTPOINT"
)

const (
	binUmount     = "umount"
	binFusermount = "fusermount"
)

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

	return exechelper.Lookup(bin, extraLookupPaths)
}

func GenerateUnmountCmd(binPath string, mountPoint string) []string {
	command := []string{binPath}
	switch filepath.Base(binPath) {
	case binFusermount:
		// linux only
		return append(command, "-z", "-u", mountPoint)
	case binUmount:
		switch runtime.GOOS {
		case "linux":
			// lazy unmount is not supported on darwin
			command = append(command, "-l")
		default:
			command = append(command, "-f")
		}
		return append(command, mountPoint)
	default:
		return nil
	}
}

func ResolveStorageCommand(bin string, args []string, remotePath, localPath string) []string {
	command := []string{bin}

	remotePath = filepath.Clean(remotePath)

	envMapping := map[string]string{
		StorageArgEnvMountpoint: localPath,
		StorageArgEnvRemotePath: remotePath,
	}

	for _, a := range args {
		command = append(command, envhelper.Expand(a, func(s, orig string) string {
			if v, ok := envMapping[s]; ok {
				return v
			}

			return orig
		}))
	}

	return command
}
