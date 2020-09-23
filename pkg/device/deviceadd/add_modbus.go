// +build !nodev,!nodev_modbus

package deviceadd

import (
	// Add modbus support
	_ "arhat.dev/arhat/pkg/device/modbus"
)
