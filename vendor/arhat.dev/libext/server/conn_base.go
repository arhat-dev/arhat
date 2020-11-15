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

package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/pkg/log"

	"arhat.dev/libext/codec"
)

func newBaseConnectionManager(
	ctx context.Context,
	logger log.Interface,
	addr net.Addr,
	handleFunc netConnectionHandleFunc,
) *baseConnectionManager {
	return &baseConnectionManager{
		ctx:    ctx,
		logger: logger,
		addr:   addr,

		handleNewConn: handleFunc,

		mu: new(sync.RWMutex),
	}
}

type baseConnectionManager struct {
	ctx    context.Context
	logger log.Interface
	addr   net.Addr

	handleNewConn netConnectionHandleFunc

	mu *sync.RWMutex
}

func (m *baseConnectionManager) validateConnection(
	conn io.Reader,
) (_ arhatgopb.ExtensionType, _ string, _ codec.Interface, err error) {
	c, ok := codec.Get(arhatgopb.CODEC_JSON)
	if !ok {
		return 0, "", nil, fmt.Errorf("json codec not found")
	}

	dec := c.NewDecoder(conn)

	initialMsg := new(arhatgopb.Msg)
	err = dec.Decode(initialMsg)
	if err != nil {
		return 0, "", nil, fmt.Errorf("invalid initial message: %w", err)
	}

	if initialMsg.Kind != arhatgopb.MSG_REGISTER {
		return 0, "", nil, fmt.Errorf("initial message is not register message")
	}

	regMsg := new(arhatgopb.RegisterMsg)
	err = c.Unmarshal(initialMsg.Payload, regMsg)
	if err != nil {
		return 0, "", nil, fmt.Errorf("failed to unmarshal register message: %w", err)
	}

	c, ok = codec.Get(regMsg.Codec)
	if !ok {
		return 0, "", nil, fmt.Errorf("codec %q not supported", regMsg.Codec.String())
	}

	return regMsg.ExtensionType, regMsg.Name, c, nil
}
