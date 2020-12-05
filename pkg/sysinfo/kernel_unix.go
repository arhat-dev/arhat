// +build !nosysinfo
// +build darwin linux freebsd openbsd netbsd dragonfly solaris aix

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

package sysinfo

import (
	"strings"

	"golang.org/x/sys/unix"
)

// nolint:unconvert
func GetKernelVersion() string {
	var uname unix.Utsname
	_ = unix.Uname(&uname)

	buf := make([]byte, len(uname.Release))
	for i, b := range uname.Release {
		buf[i] = byte(b)
	}

	kernelVersion := string(buf[:]) // nolint:gocritic
	if i := strings.Index(kernelVersion, "\x00"); i != -1 {
		kernelVersion = kernelVersion[:i]
	}

	return kernelVersion
}
