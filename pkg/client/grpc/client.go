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

package grpc

import (
	"context"
	"fmt"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
)

var (
	defaultCallOptions = []grpc.CallOption{
		grpc.WaitForReady(true),
	}
)

type Client struct {
	*clientutil.BaseClient

	serverAddress string
	dialOpts      []grpc.DialOption
	conn          *grpc.ClientConn
	client        aranyagopb.ConnectivityClient
	syncClient    aranyagopb.Connectivity_SyncClient

	mu *sync.RWMutex
}

func NewGRPCClient(agent types.Agent, config *conf.ConnectivityGRPC) (types.ConnectivityClient, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithAuthority(config.Endpoint),
	}

	tlsConfig, err := config.TLS.GetTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create client tls config: %w", err)
	}

	if tlsConfig != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	maxPayloadSize := config.MaxPayloadSize
	if maxPayloadSize <= 0 {
		maxPayloadSize = aranyagoconst.MaxGRPCDataSize
	}

	c := &Client{
		serverAddress: config.Endpoint,
		dialOpts:      dialOpts,

		mu: new(sync.RWMutex),
	}

	c.BaseClient, err = clientutil.NewBaseClient(agent, maxPayloadSize)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Connect(dialCtx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := grpc.DialContext(dialCtx, c.serverAddress, c.dialOpts...)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = aranyagopb.NewConnectivityClient(conn)

	return nil
}

func (c *Client) Start(ctx context.Context) error {
	if err := func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.syncClient != nil {
			return clientutil.ErrClientAlreadyConnected
		}

		syncClient, err := c.client.Sync(ctx, defaultCallOptions...)
		if err != nil {
			return err
		}
		c.syncClient = syncClient
		return nil
	}(); err != nil {
		return err
	}

	cmdCh := make(chan *aranyagopb.Cmd, 1)
	go func() {
		for {
			cmd, err := c.syncClient.Recv()
			if err != nil {
				close(cmdCh)

				s, _ := status.FromError(err)
				switch s.Code() {
				case codes.Canceled, codes.OK:
				default:
					c.Log.I("exception happened when client recv", log.Error(err))
				}

				return
			}

			cmdCh <- cmd
		}
	}()

	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.syncClient = nil
	}()

	for {
		select {
		case <-c.syncClient.Context().Done():
			// disconnected from cloud controller
			return c.syncClient.Context().Err()
		case <-ctx.Done():
			// leaving
			return nil
		case cmd, more := <-cmdCh:
			if !more {
				return clientutil.ErrCmdRecvClosed
			}
			c.HandleCmd(cmd)
		}
	}
}

func (c *Client) PostMsg(msg *aranyagopb.Msg) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.syncClient == nil {
		return clientutil.ErrClientNotConnected
	}

	return c.syncClient.Send(msg)
}

func (c *Client) Close() error {
	return c.OnClose(func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.syncClient != nil {
			_ = c.syncClient.CloseSend()
		}

		if c.conn != nil {
			return c.conn.Close()
		}

		return nil
	})
}
