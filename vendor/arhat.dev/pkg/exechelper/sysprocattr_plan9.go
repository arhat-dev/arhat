package exechelper

import "syscall"

func getSysProcAttr(setsid bool) *syscall.SysProcAttr {
	if setsid {
		return nil
	}

	return &syscall.SysProcAttr{}
}
