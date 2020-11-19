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
	"context"
	"fmt"

	"arhat.dev/arhat/pkg/types"
)

type (
	ConfigFactoryFunc func() interface{}
	FactoryFunc       func(
		ctx context.Context,
		handleCmd types.AgentCmdHandleFunc,
		cfg interface{},
	) (types.ConnectivityClient, error)
)

type factory struct {
	newConfig ConfigFactoryFunc
	newClient FactoryFunc
}

var supportedMethods = make(map[string]*factory)

func Register(name string, newConfig ConfigFactoryFunc, newClient FactoryFunc) {
	supportedMethods[name] = &factory{
		newConfig: newConfig,
		newClient: newClient,
	}
}

func NewConfig(name string) (interface{}, error) {
	newConfig, ok := supportedMethods[name]
	if !ok {
		return nil, fmt.Errorf("unsupported connectivity method: %s", name)
	}

	return newConfig.newConfig(), nil
}

func NewClient(
	ctx context.Context,
	name string,
	handleCmd types.AgentCmdHandleFunc,
	config interface{},
) (types.ConnectivityClient, error) {
	newClient, ok := supportedMethods[name]
	if !ok {
		return nil, fmt.Errorf("unsupported connectivity method: %s", name)
	}

	return newClient.newClient(ctx, handleCmd, config)
}
