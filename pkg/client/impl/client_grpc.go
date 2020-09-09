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

package impl

import (
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"context"
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
)

var (
	defaultCallOptions = []grpc.CallOption{
		grpc.WaitForReady(true),
	}
)

type GRPCClient struct {
	*baseClient
	serverAddress string
	dialOpts      []grpc.DialOption
	conn          *grpc.ClientConn
	client        aranyagopb.ConnectivityClient
	syncClient    aranyagopb.Connectivity_SyncClient
}

func NewGRPCClient(agent types.Agent, config *conf.ArhatGRPCConfig) (types.AgentConnectivity, error) {
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

	return &GRPCClient{
		baseClient:    newBaseClient(agent, aranyagoconst.MaxGRPCDataSize),
		serverAddress: config.Endpoint,
		dialOpts:      dialOpts,
	}, nil
}

func (c *GRPCClient) Connect(dialCtx context.Context) error {
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

func (c *GRPCClient) Start(ctx context.Context) error {
	if err := func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.syncClient != nil {
			return ErrClientAlreadyConnected
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
					c.log.I("exception happened when client recv", log.Error(err))
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
				return ErrCmdRecvClosed
			}
			c.parent.HandleCmd(cmd)
		}
	}
}

func (c *GRPCClient) PostMsg(msg *aranyagopb.Msg) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.syncClient == nil {
		return ErrClientNotConnected
	}

	return c.syncClient.Send(msg)
}

func (c *GRPCClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.syncClient != nil {
		_ = c.syncClient.CloseSend()
	}

	c.exit()

	return c.conn.Close()
}
