// +build !nodev,!nodev_gpio

package connectivityadd

import (
	// Add gpio support
	_ "arhat.dev/arhat/pkg/connectivity/gpio"
)
