// +build !nosysinfo
// +build solaris netbsd

package sysinfo

import "golang.org/x/sys/unix"

func GetTotalDiskSpace() uint64 {
	fs := unix.Statvfs_t{}
	err := unix.Statvfs("/", &fs)
	if err != nil {
		return 0
	}

	return uint64(fs.Blocks) * uint64(fs.Bsize)
}
