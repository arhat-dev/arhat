// +build !nodev,!nodev_modbus

package connectivityadd

import (
	// Add modbus support
	_ "arhat.dev/arhat/pkg/connectivity/modbus"
)
