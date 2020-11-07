// +build !solaris,!aix,!windows,!plan9

package pipenet

import (
	"syscall"
)

const syscallRead = syscall.SYS_READ

func mkfifo(path string, perm uint32) error {
	return syscall.Mkfifo(path, perm)
}
