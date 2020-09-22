// +build !nocoap

package conf

import (
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/confhelper"
)

func init() {
	RegisterConnectivityConfig("coap", func() interface{} {
		return &ConnectivityCoAP{
			ConnectivityCommonConfig: ConnectivityCommonConfig{
				Endpoint:       "",
				MaxPayloadSize: aranyagoconst.MaxCoAPDataSize,
				TLS:            confhelper.TLSConfig{},
			},
			PathNamespace: "",
			Transport:     "tcp",
			URIQueries:    make(map[string]string),
			Keepalive:     60,
		}
	})
}

type ConnectivityCoAP struct {
	ConnectivityCommonConfig `json:",inline" yaml:",inline"`

	PathNamespace string            `json:"pathNamespace" yaml:"pathNamespace"`
	Transport     string            `json:"transport" yaml:"transport"`
	URIQueries    map[string]string `json:"uriQueries" yaml:"uriQueries"`
	Keepalive     int32             `json:"keepalive" yaml:"keepalive"`
}
