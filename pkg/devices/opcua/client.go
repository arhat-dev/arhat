// +build !nodev_opcua,!nodev

package opcua

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

func init() {
	types.RegisterDeviceConnectivityFactory("opcua", aranyagopb.DEVICE_CONNECTIVITY_MODE_CLIENT, nil)
}
