// +build !noclient_grpc

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
	"io"
	"sync"
	"sync/atomic"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/aranya-proto/aranyagopb/rpcpb"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/tlshelper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/types"
)

func init() {
	client.Register("grpc",
		func() interface{} {
			return &Config{
				CommonConfig: clientutil.CommonConfig{
					Endpoint:       "",
					MaxPayloadSize: aranyagoconst.MaxGRPCDataSize,
					TLS:            tlshelper.TLSConfig{},
				},
			}
		},
		NewGRPCClient,
	)
}

type Client struct {
	*clientutil.BaseClient

	serverAddress string
	dialOpts      []grpc.DialOption
	conn          *grpc.ClientConn
	client        rpcpb.EdgeDeviceClient

	syncClientStore *atomic.Value

	mu *sync.RWMutex
}

func NewGRPCClient(
	ctx context.Context,
	handleCmd types.AgentCmdHandleFunc,
	cfg interface{},
) (client.Interface, error) {
	config, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected non grpc config")
	}

	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithAuthority(config.Endpoint),
	}

	tlsConfig, err := config.TLS.GetTLSConfig(false)
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

		syncClientStore: new(atomic.Value),

		mu: new(sync.RWMutex),
	}

	c.BaseClient, err = clientutil.NewBaseClient(ctx, handleCmd, maxPayloadSize)
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
	c.client = rpcpb.NewEdgeDeviceClient(conn)

	return nil
}

func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	// check if connected before
	if prevClient, ok := c.syncClientStore.Load().(rpcpb.EdgeDevice_SyncClient); ok && prevClient != nil {
		c.mu.Unlock()
		return clientutil.ErrClientAlreadyConnected
	}

	syncClient, err := c.client.Sync(ctx, grpc.WaitForReady(true))
	if err != nil {
		c.mu.Unlock()
		return err
	}
	c.syncClientStore.Store(syncClient)
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.syncClientStore.Store((*rpcpb.EdgeDevice_SyncClient)(nil))
		c.mu.Unlock()
	}()

	cmdCh := make(chan []byte, 1)
	go func() {
		cmd := new(aranyagopb.Cmd)
		for {
			cmd.Reset()

			err := syncClient.RecvMsg(cmd)
			if err != nil {
				close(cmdCh)

				if err == io.EOF {
					return
				}

				s, _ := status.FromError(err)
				switch s.Code() {
				case codes.Canceled, codes.OK:
				default:
					c.Log.I("exception happened when client recv", log.Error(err))
				}

				return
			}

			data, err := cmd.Marshal()
			if err != nil {
				c.Log.I("invalid cmd", log.Error(err))
				return
			}

			select {
			case <-syncClient.Context().Done():
				// disconnected from cloud controller
				return
			case <-ctx.Done():
				// leaving
				return
			case cmdCh <- data:
			}
		}
	}()

	for {
		select {
		case <-syncClient.Context().Done():
			// disconnected from cloud controller
			return syncClient.Context().Err()
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
	client, ok := c.syncClientStore.Load().(rpcpb.EdgeDevice_SyncClient)
	if !ok || client == nil {
		return clientutil.ErrClientNotConnected
	}

	return client.Send(msg)
}

func (c *Client) Close() error {
	return c.OnClose(func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		client, ok := c.syncClientStore.Load().(rpcpb.EdgeDevice_SyncClient)
		if ok && client == nil {
			_ = client.CloseSend()
		}

		if c.conn != nil {
			return c.conn.Close()
		}

		return nil
	})
}
