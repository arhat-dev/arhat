// +build !nogrpc

package conf

import (
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/confhelper"
)

func init() {
	RegisterConnectivityConfig("grpc", func() interface{} {
		return &ConnectivityGRPC{
			ConnectivityCommonConfig: ConnectivityCommonConfig{
				Endpoint:       "",
				MaxPayloadSize: aranyagoconst.MaxGRPCDataSize,
				TLS:            confhelper.TLSConfig{},
			},
		}
	})
}

type ConnectivityGRPC struct {
	ConnectivityCommonConfig `json:",inline" yaml:",inline"`
}
