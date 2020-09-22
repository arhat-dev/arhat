// +build !nomqtt

package clientadd

import (
	// Add mqtt support
	_ "arhat.dev/arhat/pkg/client/mqtt"
)
