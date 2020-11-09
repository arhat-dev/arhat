package grpc

import (
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/tlshelper"

	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/client/clientutil"
)

func init() {
	client.Register("grpc",
		func() interface{} {
			return &ConnectivityGRPC{
				ConnectivityCommonConfig: clientutil.ConnectivityCommonConfig{
					Endpoint:       "",
					MaxPayloadSize: aranyagoconst.MaxGRPCDataSize,
					TLS:            tlshelper.TLSConfig{},
				},
			}
		},
		NewGRPCClient,
	)
}
