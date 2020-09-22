package mqtt

import (
	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
)

func init() {
	client.RegisterConnectivityConfig("mqtt",
		func(agent types.Agent, clientConfig interface{}) (types.ConnectivityClient, error) {
			return NewMQTTClient(agent, clientConfig.(*conf.ConnectivityMQTT))
		},
	)
}
