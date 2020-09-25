// +build !nodev,!nodev_mqtt

package connectivityadd

import (
	// Add mqtt support
	_ "arhat.dev/arhat/pkg/connectivity/mqtt"
)
