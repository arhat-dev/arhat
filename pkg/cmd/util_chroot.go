// +build !windows,!zos

package cmd

import (
	"syscall"
)

func chroot(dir string) error {
	return syscall.Chroot(dir)
}
