// +build !nodev,!nodev_gpio

package deviceadd

import (
	// Add gpio support
	_ "arhat.dev/arhat/pkg/device/gpio"
)
