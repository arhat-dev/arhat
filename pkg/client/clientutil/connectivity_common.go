package clientutil

import "arhat.dev/pkg/tlshelper"

type ConnectivityCommonConfig struct {
	Endpoint       string              `json:"endpoint" yaml:"endpoint"`
	MaxPayloadSize int                 `json:"maxPayloadSize" yaml:"maxPayloadSize"`
	TLS            tlshelper.TLSConfig `json:"tls" yaml:"tls"`
}
