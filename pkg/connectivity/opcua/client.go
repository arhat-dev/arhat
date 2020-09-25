package opcua

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/connectivity"
)

func init() {
	connectivity.Register("opcua", aranyagopb.CONNECTIVITY_MODE_CLIENT, nil)
}
