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

package clientutil

import (
	"context"
	"errors"
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"

	"arhat.dev/arhat/pkg/types"
)

var (
	ErrClientAlreadyConnected = errors.New("client already connected")
	ErrClientNotConnected     = errors.New("client not connected")
	ErrCmdRecvClosed          = errors.New("cmd recv closed")
)

func NewBaseClient(
	ctx context.Context,
	handleCmd types.AgentCmdHandleFunc,
	maxPayloadSize int,
) (*BaseClient, error) {
	ctx, cancel := context.WithCancel(ctx)

	if maxPayloadSize-aranyagopb.EmptyMsgSize <= 0 {
		cancel()
		return nil, fmt.Errorf("maxPayloadSize must be greater than %d", aranyagopb.EmptyMsgSize)
	}

	return &BaseClient{
		ctx:  ctx,
		exit: cancel,

		Log:            log.Log.WithName("client"),
		maxPayloadSize: maxPayloadSize,

		handleCmd: handleCmd,
	}, nil
}

type BaseClient struct {
	ctx  context.Context
	exit context.CancelFunc

	maxPayloadSize int

	Log       log.Interface
	handleCmd types.AgentCmdHandleFunc
}

func (b *BaseClient) HandleCmd(cmd *aranyagopb.Cmd) {
	b.handleCmd(cmd)
}

func (b *BaseClient) OnClose(doExit func() error) error {
	b.exit()

	if doExit != nil {
		return doExit()
	}

	return nil
}

func (b *BaseClient) Context() context.Context {
	return b.ctx
}

func (b *BaseClient) MaxPayloadSize() int {
	return b.maxPayloadSize
}
