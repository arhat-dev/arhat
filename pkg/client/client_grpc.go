// +build grpc

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
	"arhat.dev/arhat/pkg/client/impl"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
)

func NewClient(agent types.Agent, clientConfig *conf.ArhatConnectivityMethods) (types.AgentConnectivity, error) {
	return impl.NewGRPCClient(agent, &clientConfig.GRPCConfig)
}
