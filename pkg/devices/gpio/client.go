// +build !nodev_gpio,!nodev

package gpio

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

func init() {
	types.RegisterDeviceConnectivityFactory("http", aranyagopb.DEVICE_CONNECTIVITY_MODE_CLIENT, nil)
}
