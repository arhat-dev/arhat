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

	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

type (
	ConfigFactoryFunc func() interface{}
	FactoryFunc       func(
		ctx context.Context,
		handleCmd types.AgentCmdHandleFunc,
		cfg interface{},
	) (Interface, error)
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

type Interface interface {
	// Context of this client
	Context() context.Context

	// Connect to server/broker
	Connect(dialCtx context.Context) error

	// Start internal logic to get prepared for communicating with aranya
	// usually send online state message
	Start(appCtx context.Context) error

	// PostMsg to aranya
	PostMsg(msg *aranyagopb.Msg) error

	// Close this client
	Close() error

	// MaxPayloadSize of a single message for this client
	MaxPayloadSize() int
}

func NewClient(
	ctx context.Context,
	name string,
	handleCmd types.AgentCmdHandleFunc,
	config interface{},
) (Interface, error) {
	newClient, ok := supportedMethods[name]
	if !ok {
		return nil, fmt.Errorf("unsupported connectivity method: %s", name)
	}

	return newClient.newClient(ctx, handleCmd, config)
}
