package exechelper

import (
	"syscall"
)

func getSysProcAttr(tty bool) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
