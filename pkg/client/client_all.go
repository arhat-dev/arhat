// +build !grpc,!mqtt,!coap

/*
Copyright 2019 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"fmt"
	"strings"

	"arhat.dev/arhat/pkg/client/impl"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
)

func NewClient(agent types.Agent, clientConfig *conf.ArhatConnectivityMethods) (types.AgentConnectivity, error) {
	switch strings.ToLower(clientConfig.Method) {
	case aranyagoconst.ConnectivityMethodMQTT:
		return impl.NewMQTTClient(agent, &clientConfig.MQTTConfig)
	case aranyagoconst.ConnectivityMethodGRPC:
		return impl.NewGRPCClient(agent, &clientConfig.GRPCConfig)
	case aranyagoconst.ConnectivityMethodCoAP:
		return impl.NewCoAPClient(agent, &clientConfig.CoAPConfig)
	}

	return nil, fmt.Errorf("connectivity config not provided for method %s", clientConfig.Method)
}
