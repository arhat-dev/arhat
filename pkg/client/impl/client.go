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

package impl

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"

	"arhat.dev/arhat/pkg/types"
)

var (
	ErrClientAlreadyConnected = errors.New("client already connected")
	ErrClientNotConnected     = errors.New("client not connected")
	ErrCmdRecvClosed          = errors.New("cmd recv closed")
)

func newBaseClient(agent types.Agent, maxPayloadSize int) *baseClient {
	ctx, cancel := context.WithCancel(agent.Context())

	maxPayloadSize -= aranyagopb.EmptyMsgSize
	if maxPayloadSize <= 0 {
		panic(fmt.Errorf("maxPayloadSize must be greater than %d", aranyagopb.EmptyMsgSize))
	}

	return &baseClient{
		ctx:  ctx,
		exit: cancel,

		log:            log.Log.WithName("client"),
		maxPayloadSize: maxPayloadSize,
		mu:             new(sync.RWMutex),

		parent: agent,
	}
}

type baseClient struct {
	ctx  context.Context
	exit context.CancelFunc

	log            log.Interface
	maxPayloadSize int
	mu             *sync.RWMutex

	parent types.Agent
}

func (b *baseClient) Context() context.Context {
	return b.ctx
}

func (b *baseClient) MaxPayloadSize() int {
	return b.maxPayloadSize
}
