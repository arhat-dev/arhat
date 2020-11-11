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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func CleanupPodData(podDir, remoteVolumeDir, tmpfsVolumeDir string, unmount func(path string) error) error {
	err := RemoveAllOneFilesystem(podDir)
	for n := 0; n < 5 && err != nil; n++ {
		dirs, _ := ioutil.ReadDir(remoteVolumeDir)
		for _, dir := range dirs {
			err = unmount(filepath.Join(remoteVolumeDir, dir.Name()))
		}

		if err == nil {
			dirs, _ = ioutil.ReadDir(tmpfsVolumeDir)
			for _, dir := range dirs {
				// TODO: unmount tmpfs
				_ = filepath.Join(tmpfsVolumeDir, dir.Name())
			}
		}

		time.Sleep(5 * time.Second)
		if err == nil {
			err = RemoveAllOneFilesystem(podDir)
		}
	}

	return err
}

func ResolveHostPathMountSource(
	path, podUID, volName string,
	remote bool,
	getRemoteVolumeDir func(podUID, volName string) string,
	getTmpfsVolumeDir func(podUID, volName string) string,
) (string, error) {
	var mountSource string
	switch {
	case path != "":
		if !remote {
			return path, nil
		}

		// mount remote volume
		mountSource = getRemoteVolumeDir(podUID, volName)
		if err := os.MkdirAll(mountSource, 0750); err != nil && !os.IsExist(err) {
			return "", err
		}

		return mountSource, nil
	case path == "" && !remote:
		// make an empty dir (pretend to be tmpfs)
		mountSource = getTmpfsVolumeDir(podUID, volName)
		if err := os.MkdirAll(mountSource, 0750); err != nil && !os.IsExist(err) {
			return "", err
		}

		return mountSource, nil
	default:
		return "", fmt.Errorf("unsupported options")
	}
}
