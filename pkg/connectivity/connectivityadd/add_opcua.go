// +build !nodev,!nodev_opcua

package connectivityadd

import (
	// Add opc-ua support
	_ "arhat.dev/arhat/pkg/connectivity/opcua"
)
