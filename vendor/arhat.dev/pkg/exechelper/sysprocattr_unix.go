// +build !windows,!js,!plan9

package exechelper

import (
	"syscall"
)

func getSysProcAttr(setsid bool) *syscall.SysProcAttr {
	// https://github.com/creack/pty/issues/35#issuecomment-147947212
	// do not Setpgid if already Setsid
	if setsid {
		return nil
	}

	return &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
}
