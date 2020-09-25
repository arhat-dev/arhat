package http

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/connectivity"
)

func init() {
	connectivity.Register("http", aranyagopb.CONNECTIVITY_MODE_CLIENT, nil)
}
