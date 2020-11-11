// +build !nostorage_general
// +build !windows

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

package general

import (
	"fmt"

	"ext.arhat.dev/runtimeutil/storageutil"
)

func init() {
	storageutil.Register(
		"general",
		func(config interface{}) (storageutil.Interface, error) {
			return New(config)
		},
		func() interface{} {
			return &Config{}
		},
	)
}

func New(cfg interface{}) (*Driver, error) {
	config, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config")
	}

	return &Driver{config: config}, nil
}

type Config struct {
	Command string   `json:"command" yaml:"command"`
	Args    []string `json:"args" yaml:"args"`

	Fuse bool `json:"fuse" yaml:"fuse"`
}

type Driver struct {
	config *Config
}

func (d *Driver) GetMountCmd(remotePath, mountPoint string) []string {
	return storageutil.ResolveStorageCommand(
		d.config.Command, d.config.Args, remotePath, mountPoint,
	)
}

func (d *Driver) GetUnmountCmd(mountPoint string) []string {
	umountBin, err := storageutil.LookupUnmountUtil(nil, d.config.Fuse)
	if err != nil {
		return nil
	}

	return storageutil.GenerateUnmountCmd(umountBin, mountPoint)
}
