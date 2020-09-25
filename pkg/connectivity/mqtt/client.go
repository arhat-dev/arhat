package mqtt

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/connectivity"
)

func init() {
	connectivity.Register("mqtt", aranyagopb.CONNECTIVITY_MODE_CLIENT, nil)
}
