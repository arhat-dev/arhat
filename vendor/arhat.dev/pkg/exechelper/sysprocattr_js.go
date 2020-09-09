package exechelper

import "syscall"

func getSysProcAttr(setsid bool) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
