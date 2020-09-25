package coap

import (
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/confhelper"

	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/client/clientutil"
)

func init() {
	client.Register("coap",
		func() interface{} {
			return &ConnectivityCoAP{
				ConnectivityCommonConfig: clientutil.ConnectivityCommonConfig{
					Endpoint:       "",
					MaxPayloadSize: aranyagoconst.MaxCoAPDataSize,
					TLS:            confhelper.TLSConfig{},
				},
				PathNamespace: "",
				Transport:     "tcp",
				URIQueries:    make(map[string]string),
				Keepalive:     60,
			}
		},
		NewCoAPClient,
	)
}
