// +build !windows,!zos,!plan9,!js

package main

import (
	"syscall"
)

func chroot(dir string) error {
	return syscall.Chroot(dir)
}
