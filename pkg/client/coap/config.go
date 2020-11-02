package coap

import (
	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/conf"
)

type ConnectivityCoAP struct {
	clientutil.ConnectivityCommonConfig `json:",inline" yaml:",inline"`

	PathNamespaceFrom conf.ValueFromSpec `json:"pathNamespaceFrom" yaml:"pathNamespaceFrom"`
	Transport         string             `json:"transport" yaml:"transport"`
	URIQueries        map[string]string  `json:"uriQueries" yaml:"uriQueries"`
	Keepalive         int32              `json:"keepalive" yaml:"keepalive"`
}
