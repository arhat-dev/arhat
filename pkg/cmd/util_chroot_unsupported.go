// +build windows zos plan9 js

package cmd

import "fmt"

func chroot(_ string) error {
	return fmt.Errorf("chroot is not supported")
}
