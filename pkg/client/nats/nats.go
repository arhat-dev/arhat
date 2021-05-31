// +build !noclient_nats

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

package nats

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"

	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/types"
)

func init() {
	client.Register("nats",
		func() interface{} {
			return &Config{
				CommonConfig: clientutil.CommonConfig{
					MaxPayloadSize: 1024 * 1024,
				},
				PingInterval:     nats.DefaultPingInterval,
				AckWait:          15 * time.Second,
				MaxPendingPubAck: 16,
			}
		},
		NewClient,
	)
}

func NewClient(
	ctx context.Context,
	handleCmd types.AgentCmdHandleFunc,
	cfg interface{},
) (_ client.Interface, err error) {
	config, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected non nats config")
	}

	if config.Endpoint == "" {
		return nil, fmt.Errorf("invalid empty nats streaming server endpoint: %w", err)
	}

	if config.SubjectNamespace == "" {
		return nil, fmt.Errorf("invalid empty subject namespace: %w", err)
	}

	baseClient, err := clientutil.NewBaseClient(ctx, handleCmd, config.MaxPayloadSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create base client: %w", err)
	}

	tlsConfig, err := config.TLS.GetTLSConfig(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get tls config: %w", err)
	}

	disconnSig := make(chan error, 1)

	nats.UserCredentials("")
	options := &nats.Options{
		Servers:     []string{config.Endpoint},
		NoRandomize: false,
		NoEcho:      true,
		Name:        config.ClientID,

		// auth method is handled later
		Nkey:        "",
		UserJWT:     nil,
		SignatureCB: nil,

		User:         config.Username,
		Password:     config.Password,
		Token:        config.Token,
		TokenHandler: nil,

		Verbose:   false, // disable +OK/-ERR ack
		Pedantic:  false,
		Secure:    tlsConfig != nil && tlsConfig.InsecureSkipVerify,
		TLSConfig: tlsConfig,

		// disable auto reconnect
		AllowReconnect:         false,
		MaxReconnect:           0,
		ReconnectWait:          0,
		CustomReconnectDelayCB: nil,
		ReconnectJitter:        0,
		ReconnectJitterTLS:     0,
		ReconnectedCB:          nil,
		ReconnectBufSize:       0,

		CustomDialer:   nil, // set when connect
		Timeout:        0,   // set when connect
		DrainTimeout:   nats.DefaultDrainTimeout,
		FlusherTimeout: 0, // no flush delay

		PingInterval: config.PingInterval,
		MaxPingsOut:  nats.DefaultMaxPingOut,
		DisconnectedErrCB: func(c *nats.Conn, e error) {
			select {
			case <-disconnSig:
			default:
				disconnSig <- e
				close(disconnSig)
			}
		},

		ClosedCB:            nil,
		DiscoveredServersCB: nil, // no discovery
		AsyncErrorCB:        nil, // no error handling regarding message sub
		SubChanLen:          0,   // no sync sub

		UseOldRequestStyle:          false,
		NoCallbacksAfterClientClose: true,
	}

	if config.NKey != "" {
		nkeyBytes, err2 := base64.StdEncoding.DecodeString(config.NKey)
		if err2 != nil {
			return nil, fmt.Errorf("invalid base64 encoded nkey: %w", err2)
		}

		kp, err2 := jwt.ParseDecoratedNKey(nkeyBytes)
		if err2 != nil {
			return nil, fmt.Errorf("invalid nkey: %w", err2)
		}

		if config.JWT != "" {
			jwtBytes, err2 := base64.StdEncoding.DecodeString(config.JWT)
			if err2 != nil {
				return nil, fmt.Errorf("invalid jwt: %w", err2)
			}

			userJWT, err2 := jwt.ParseDecoratedJWT(jwtBytes)
			if err2 != nil {
				return nil, fmt.Errorf("invalid user jwt: %w", err2)
			}

			options.UserJWT = func() (string, error) {
				return userJWT, nil
			}

			options.SignatureCB = kp.Sign
		} else {
			pub, err := kp.PublicKey()
			if err != nil {
				return nil, fmt.Errorf("failed to get public key of nkey: %w", err)
			}

			if !nkeys.IsValidPublicUserKey(pub) {
				return nil, fmt.Errorf("invalid nkey user seed")
			}

			options.Nkey = pub
			options.SignatureCB = kp.Sign
		}
	}

	maxPendingPubAck := config.MaxPendingPubAck
	if maxPendingPubAck == 0 {
		maxPendingPubAck = 16
	}

	ackWait := config.AckWait
	if ackWait == 0 {
		ackWait = 15 * time.Second
	}

	client := &Client{
		ctx: ctx,

		BaseClient: baseClient,

		onlineMsgBytes:  nil,
		offlineMsgBytes: nil,

		ackWait:          ackWait,
		maxPendingPubAck: maxPendingPubAck,
		clusterID:        config.ClusterID,

		options: options,

		disconnSig: disconnSig,
	}

	client.onlineMsgBytes, client.offlineMsgBytes = clientutil.CreateOnlineOfflineMessage(config.ClientID)

	client.cmdSubTopic, client.msgPubTopic, client.statePubTopic = aranyagoconst.NatsTopics(config.SubjectNamespace)

	return client, nil
}

type Client struct {
	ctx context.Context

	*clientutil.BaseClient

	onlineMsgBytes  []byte
	offlineMsgBytes []byte

	cmdSubTopic   string
	msgPubTopic   string
	statePubTopic string

	ackWait          time.Duration
	maxPendingPubAck int
	clusterID        string

	options *nats.Options
	nc      *nats.Conn
	client  stan.Conn

	disconnSig chan error
}

func (c *Client) Connect(dialCtx context.Context) error {
	deadline, ok := dialCtx.Deadline()
	if ok {
		c.options.CustomDialer = &net.Dialer{
			Deadline: deadline,
		}
		c.options.Timeout = time.Until(deadline)
	}

	var err error
	c.nc, err = c.options.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to nats server: %w", err)
	}

	c.client, err = stan.Connect(c.clusterID, c.options.Name,
		stan.NatsConn(c.nc),
		stan.MaxPubAcksInflight(c.maxPendingPubAck),
		stan.PubAckWait(c.ackWait),
	)
	if err != nil {
		c.nc.Close()
		return fmt.Errorf("failed to create nats streaming client: %w", err)
	}

	return nil
}

func (c *Client) Start(appCtx context.Context) error {
	sub, err := c.client.Subscribe(c.cmdSubTopic, func(msg *stan.Msg) {
		c.HandleCmd(msg.Data)
	}, stan.StartAt(pb.StartPosition_NewOnly), stan.AckWait(c.ackWait))
	if err != nil {
		return fmt.Errorf("failed to subscribe cmd topic: %w", err)
	}

	defer func() {
		// publish offline message, best effort
		_ = c.client.Publish(c.statePubTopic, c.offlineMsgBytes)

		_ = sub.Unsubscribe()
		_ = sub.Close()

		_ = c.client.Close()
		c.nc.Close()
	}()

	// publish online message
	err = c.client.Publish(c.msgPubTopic, c.onlineMsgBytes)
	if err != nil {
		return fmt.Errorf("failed to publish initial online message: %w", err)
	}

	select {
	case <-c.ctx.Done():
		return nil
	case <-appCtx.Done():
		return nil
	case err := <-c.disconnSig:
		return err
	}
}

func (c *Client) postData(data []byte) error {
	guid, err := c.client.PublishAsync(c.msgPubTopic, data, func(guid string, err error) {
		if err == nil {
			return
		}

		if c.nc.IsClosed() {
			// underlay nats client closed
			return
		}

		select {
		case <-c.disconnSig:
			// disconnected
			return
		default:
			// alive
		}

		_ = c.postData(data)
	})
	_ = guid

	return err
}

func (c *Client) PostMsg(msg *aranyagopb.Msg) error {
	data, err := msg.Marshal()
	if err != nil {
		return err
	}

	return c.postData(data)
}

func (c *Client) Close() error {
	return c.OnClose(func() error {
		if c.client != nil {
			c.client.Close()
		}

		if c.nc != nil {
			c.nc.Close()
		}
		return nil
	})
}
