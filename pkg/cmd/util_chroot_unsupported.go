// +build windows zos plan9

package cmd

import "fmt"

func chroot(_ string) error {
	return fmt.Errorf("chroot is not supported")
}
