// +build !nodev,!nodev_mqtt

package deviceadd

import (
	// Add mqtt support
	_ "arhat.dev/arhat/pkg/device/mqtt"
)
