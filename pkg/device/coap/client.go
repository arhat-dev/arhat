package coap

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

func init() {
	types.RegisterDeviceConnectivityFactory("coap", aranyagopb.DEVICE_CONNECTIVITY_MODE_CLIENT, nil)
}
