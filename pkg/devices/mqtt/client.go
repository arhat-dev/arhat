// +build !nodev_mqtt,!nodev

package mqtt

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

func init() {
	types.RegisterDeviceConnectivityFactory("mqtt", aranyagopb.DEVICE_CONNECTIVITY_MODE_CLIENT, nil)
}
