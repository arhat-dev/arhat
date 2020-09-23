// +build !nodev,!nodev_opcua

package deviceadd

import (
	// Add opc-ua support
	_ "arhat.dev/arhat/pkg/device/opcua"
)
