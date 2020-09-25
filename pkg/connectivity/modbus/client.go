package modbus

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/connectivity"
)

func init() {
	connectivity.Register("modbus", aranyagopb.CONNECTIVITY_MODE_CLIENT, nil)
}
