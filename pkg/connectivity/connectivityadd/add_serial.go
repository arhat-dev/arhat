// +build !nodev,!nodev_serial

package connectivityadd

import (
	// Add serial support
	_ "arhat.dev/arhat/pkg/connectivity/serial"
)
