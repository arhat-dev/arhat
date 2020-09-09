// +build !nosysinfo

package sysinfo

import "golang.org/x/sys/unix"

func GetTotalDiskSpace() uint64 {
	fs := unix.Statfs_t{}
	err := unix.Statfs("/", &fs)
	if err != nil {
		return 0
	}

	return uint64(fs.F_blocks) * uint64(fs.F_bsize)
}
