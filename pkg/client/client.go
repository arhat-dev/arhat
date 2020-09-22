/*
Copyright 2020 The arhat.dev Authors.

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
	"fmt"

	"arhat.dev/arhat/pkg/types"
)

var supportedConnectivityClients = make(map[string]types.ConnectivityClientFactoryFunc)

func RegisterConnectivityConfig(name string, newClient types.ConnectivityClientFactoryFunc) {
	supportedConnectivityClients[name] = newClient
}

func NewConnectivityClient(name string, agent types.Agent, config interface{}) (types.ConnectivityClient, error) {
	newClient, ok := supportedConnectivityClients[name]
	if !ok {
		return nil, fmt.Errorf("unsupported connectivity method: %s", name)
	}

	return newClient(agent, config)
}
