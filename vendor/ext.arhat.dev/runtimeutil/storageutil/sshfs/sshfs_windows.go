// +build !nostorage_sshfs

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

package sshfs

import (
	"fmt"

	"ext.arhat.dev/runtimeutil/storageutil"
)

func init() {
	storageutil.Register(
		"sshfs",
		func(config interface{}) (storageutil.Interface, error) {
			return New(config)
		},
		func() interface{} {
			return new(Config)
		},
	)
}

type Config struct {
	// TBD
}

func New(cfg interface{}) (*Driver, error) {
	return nil, fmt.Errorf("not supported")
}

type Driver struct {
	// TBD
}

func (d *Driver) GetMountCmd(remotePath, mountPoint string) []string {
	return nil
}

func (d *Driver) GetUnmountCmd(mountPoint string) []string {
	return nil
}
