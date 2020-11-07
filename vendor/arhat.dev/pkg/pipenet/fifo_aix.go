package pipenet

import (
	"golang.org/x/sys/unix"
)

const syscallRead = 3

func mkfifo(path string, perm uint32) error {
	return unix.Mkfifo(path, perm)
}
