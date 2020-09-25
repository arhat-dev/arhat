package coap

import (
	"arhat.dev/arhat/pkg/client/clientutil"
)

type ConnectivityCoAP struct {
	clientutil.ConnectivityCommonConfig `json:",inline" yaml:",inline"`

	PathNamespace string            `json:"pathNamespace" yaml:"pathNamespace"`
	Transport     string            `json:"transport" yaml:"transport"`
	URIQueries    map[string]string `json:"uriQueries" yaml:"uriQueries"`
	Keepalive     int32             `json:"keepalive" yaml:"keepalive"`
}
