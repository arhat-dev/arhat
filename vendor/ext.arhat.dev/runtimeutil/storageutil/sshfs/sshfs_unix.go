// +build !nostorage_sshfs
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

package sshfs

import (
	"fmt"
	"os"
	"strings"

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
	Args []string `json:"args" yaml:"args"`
}

func New(cfg interface{}) (*Driver, error) {
	config, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config")
	}

	// validate args
	if len(config.Args) < 2 {
		return nil, fmt.Errorf("expect at least 2 args")
	}

	valid := true
	count := 0
	// first arg MUST include env ref REMOTE_PATH
	os.Expand(config.Args[0], func(s string) string {
		count++
		if s != storageutil.StorageArgEnvRemotePath {
			valid = false
		}
		return ""
	})
	if !valid || count > 1 {
		return nil, fmt.Errorf(
			"first arg invalid, must include exactly one $%s",
			storageutil.StorageArgEnvRemotePath,
		)
	}

	count = 0
	// second arg MUST include env ref for mount point
	localPath := os.Expand(config.Args[1], func(s string) string {
		count++
		if s != storageutil.StorageArgEnvMountpoint {
			valid = false
		}
		return ""
	})
	if !valid || count > 1 || localPath != "" {
		return nil, fmt.Errorf("second arg invalid, must be $%s", storageutil.StorageArgEnvMountpoint)
	}

	blacklistOptionsPrefix := []string{
		// do not allow any stdin related args
		"password_stdin",
		"slave",

		// TODO: should we not allow using any ssh config to avoid stdin being used?
		// "-F",
	}

	// other args should not contain any env ref

	for _, arg := range config.Args[2:] {
		os.Expand(arg, func(s string) string {
			valid = false
			return ""
		})

		if !valid {
			return nil, fmt.Errorf("invalid arg %s: should not contain env ref", arg)
		}

		for _, prefix := range blacklistOptionsPrefix {
			if strings.Contains(arg, prefix) {
				return nil, fmt.Errorf("option %s is not allowed", arg)
			}
		}
	}

	// ensure foreground and auto reconnect
	return &Driver{args: append(append([]string{}, config.Args...), "-f", "-o", "reconnect")}, nil
}

type Driver struct {
	args []string
}

func (d *Driver) GetMountCmd(remotePath, mountPoint string) []string {
	return storageutil.ResolveStorageCommand("sshfs", d.args, remotePath, mountPoint)
}

func (d *Driver) GetUnmountCmd(mountPoint string) []string {
	umountBin, err := storageutil.LookupUnmountUtil(nil, true)
	if err != nil {
		return nil
	}

	return storageutil.GenerateUnmountCmd(umountBin, mountPoint)
}
