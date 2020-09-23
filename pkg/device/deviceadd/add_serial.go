// +build !nodev,!nodev_serial

package deviceadd

import (
	// Add serial support
	_ "arhat.dev/arhat/pkg/device/serial"
)
