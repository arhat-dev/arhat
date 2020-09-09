// +build !nosysinfo
// +build !windows,!js,!plan9

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
