// +build !windows,!zos,!plan9,!js

package cmd

import (
	"syscall"
)

func chroot(dir string) error {
	return syscall.Chroot(dir)
}
