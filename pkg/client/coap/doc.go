package coap

import (
	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
)

func init() {
	client.RegisterConnectivityConfig("coap",
		func(agent types.Agent, clientConfig interface{}) (types.ConnectivityClient, error) {
			return NewCoAPClient(agent, clientConfig.(*conf.ConnectivityCoAP))
		},
	)
}
