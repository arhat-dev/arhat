package mqtt

import (
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/confhelper"

	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/conf"
)

func init() {
	client.RegisterConnectivityConfig("mqtt",
		func() interface{} {
			return &ConnectivityMQTT{
				ConnectivityCommonConfig: clientutil.ConnectivityCommonConfig{
					Endpoint:       "",
					MaxPayloadSize: aranyagoconst.MaxMQTTDataSize,
					TLS:            confhelper.TLSConfig{},
				},
				Version:            "3.1.1",
				Variant:            "standard",
				Transport:          "tcp",
				TopicNamespaceFrom: conf.ValueFromSpec{},
				ClientID:           "",
				Username:           "",
				Password:           "",
				Keepalive:          60,
			}
		},
		NewMQTTClient,
	)
}
