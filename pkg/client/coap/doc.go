package coap

import (
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/tlshelper"

	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/conf"
)

func init() {
	client.Register("coap",
		func() interface{} {
			return &ConnectivityCoAP{
				ConnectivityCommonConfig: clientutil.ConnectivityCommonConfig{
					Endpoint:       "",
					MaxPayloadSize: aranyagoconst.MaxCoAPDataSize,
					TLS:            tlshelper.TLSConfig{},
				},
				PathNamespaceFrom: conf.ValueFromSpec{},
				Transport:         "tcp",
				URIQueries:        make(map[string]string),
				Keepalive:         60,
			}
		},
		NewCoAPClient,
	)
}
