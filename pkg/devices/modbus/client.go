// +build !nodev_modbus,!nodev

package modbus

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

func init() {
	types.RegisterDeviceConnectivityFactory("modbus", aranyagopb.DEVICE_CONNECTIVITY_MODE_CLIENT, nil)
}
