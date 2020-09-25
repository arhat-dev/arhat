// +build !nodev,!nodev_coap

package connectivityadd

import (
	// Add coap support
	_ "arhat.dev/arhat/pkg/connectivity/coap"
)
