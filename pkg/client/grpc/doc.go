package grpc

import (
	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
)

func init() {
	client.RegisterConnectivityConfig("grpc",
		func(agent types.Agent, clientConfig interface{}) (types.ConnectivityClient, error) {
			return NewGRPCClient(agent, clientConfig.(*conf.ConnectivityGRPC))
		},
	)
}
