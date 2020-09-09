// +build !nosysinfo
// +build darwin freebsd linux

package sysinfo

import "golang.org/x/sys/unix"

// nolint:unconvert
func GetTotalDiskSpace() uint64 {
	fs := unix.Statfs_t{}
	err := unix.Statfs("/", &fs)
	if err != nil {
		return 0
	}

	return uint64(fs.Blocks) * uint64(fs.Bsize)
}
