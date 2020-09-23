// +build !nodev,!nodev_coap

package deviceadd

import (
	// Add coap support
	_ "arhat.dev/arhat/pkg/device/coap"
)
