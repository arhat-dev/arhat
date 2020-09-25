package clientutil

import "arhat.dev/pkg/confhelper"

type ConnectivityCommonConfig struct {
	Endpoint       string               `json:"endpoint" yaml:"endpoint"`
	MaxPayloadSize int                  `json:"maxPayloadSize" yaml:"maxPayloadSize"`
	TLS            confhelper.TLSConfig `json:"tls" yaml:"tls"`
}
