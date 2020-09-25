package grpc

import (
	"arhat.dev/arhat/pkg/client/clientutil"
)

type ConnectivityGRPC struct {
	clientutil.ConnectivityCommonConfig `json:",inline" yaml:",inline"`
}
