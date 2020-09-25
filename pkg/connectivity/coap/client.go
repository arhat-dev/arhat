package coap

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/connectivity"
)

func init() {
	connectivity.Register("coap", aranyagopb.CONNECTIVITY_MODE_CLIENT, nil)
}
