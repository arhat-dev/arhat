// +build !noclient_mqtt

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

package mqtt

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/tlshelper"
	"github.com/goiiot/libmqtt"

	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
)

func init() {
	client.Register("mqtt",
		func() interface{} {
			return &Config{
				CommonConfig: clientutil.CommonConfig{
					Endpoint:       "",
					MaxPayloadSize: aranyagoconst.MaxMQTTDataSize,
					TLS:            tlshelper.TLSConfig{},
				},
				Version:            "3.1.1",
				Variant:            "standard",
				Transport:          "tcp",
				TopicNamespaceFrom: conf.ValueFromSpec{},
				ClientID:           "",
				Username:           "",
				Password:           "",
				Keepalive:          60,
			}
		},
		NewMQTTClient,
	)
}

func NewMQTTClient(
	ctx context.Context,
	handleCmd types.AgentCmdHandleFunc,
	cfg interface{},
) (_ client.Interface, err error) {
	config, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected non mqtt config")
	}

	var options []libmqtt.Option
	switch config.Version {
	case "5":
		options = append(options, libmqtt.WithVersion(libmqtt.V5, false))
	case "3.1.1", "":
		options = append(options, libmqtt.WithVersion(libmqtt.V311, false))
	default:
		return nil, fmt.Errorf("unsupported mqtt version: %s", config.Version)
	}

	switch config.Transport {
	case "websocket":
		options = append(options, libmqtt.WithWebSocketConnector(0, nil))
	case "tcp", "":
		options = append(options, libmqtt.WithTCPConnector(0))
	default:
		return nil, fmt.Errorf("unsupported transport method: %s", config.Transport)
	}

	connInfo, err := config.GetConnectInfo()
	if err != nil {
		return nil, fmt.Errorf("invalid config options for mqtt connect: %w", err)
	}

	onlineMsgBytes, _ := (&aranyagopb.StateMsg{
		Kind:     aranyagopb.STATE_ONLINE,
		DeviceId: connInfo.ClientID,
	}).Marshal()
	onlineMsgBytes, _ = (&aranyagopb.Msg{
		Kind:      aranyagopb.MSG_STATE,
		Sid:       0,
		Seq:       0,
		Completed: true,
		Payload:   onlineMsgBytes,
	}).Marshal()

	willMsgBytes, _ := (&aranyagopb.StateMsg{
		Kind:     aranyagopb.STATE_OFFLINE,
		DeviceId: connInfo.ClientID,
	}).Marshal()
	willMsgBytes, _ = (&aranyagopb.Msg{
		Kind:      aranyagopb.MSG_STATE,
		Sid:       0,
		Seq:       0,
		Completed: true,
		Payload:   willMsgBytes,
	}).Marshal()

	if connInfo.TLSConfig != nil {
		options = append(options, libmqtt.WithCustomTLS(connInfo.TLSConfig))
	}

	keepalive := config.Keepalive
	if keepalive == 0 {
		// default to 60 seconds
		keepalive = 60
	}

	options = append(options, libmqtt.WithRouter(connInfo.TopicRouter))
	options = append(options, libmqtt.WithConnPacket(libmqtt.ConnPacket{
		Username:     connInfo.Username,
		Password:     connInfo.Password,
		ClientID:     connInfo.ClientID,
		Keepalive:    uint16(keepalive),
		CleanSession: true,
		IsWill:       true,
		WillTopic:    connInfo.WillPubTopic,
		WillQos:      libmqtt.Qos1,
		WillRetain:   connInfo.SupportRetain,
		WillMessage:  willMsgBytes,
	}))

	options = append(options, libmqtt.WithKeepalive(uint16(keepalive), 1.2))

	client, err := libmqtt.NewClient(options...)
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:        client,
		supportRetain: connInfo.SupportRetain,

		brokerAddress:     config.Endpoint,
		cmdSubTopicHandle: connInfo.CmdSubTopicHandle,
		cmdSubTopic:       connInfo.CmdSubTopic,
		pubTopic:          connInfo.MsgPubTopic,
		pubWillTopic:      connInfo.WillPubTopic,
		onlineMsg:         onlineMsgBytes,

		netErrCh:  make(chan error),
		connErrCh: make(chan error),
		subErrCh:  make(chan error),
	}

	c.BaseClient, err = clientutil.NewBaseClient(ctx, handleCmd, connInfo.MaxPayloadSize)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type Client struct {
	onlineMsg         []byte
	brokerAddress     string
	cmdSubTopicHandle string
	cmdSubTopic       string
	pubTopic          string
	pubWillTopic      string

	*clientutil.BaseClient
	client libmqtt.Client

	netErrCh  chan error
	connErrCh chan error
	subErrCh  chan error

	exited        int32
	supportRetain bool
}

func (c *Client) Connect(dialCtx context.Context) error {
	dialOpts := []libmqtt.Option{
		libmqtt.WithAutoReconnect(false),
		libmqtt.WithConnHandleFunc(c.handleConn(c.Context().Done())),
		libmqtt.WithSubHandleFunc(c.handleSub(c.Context().Done())),
		libmqtt.WithPubHandleFunc(c.handlePub),
		libmqtt.WithNetHandleFunc(c.handleNet),
	}

	dd, ok := dialCtx.Deadline()
	if ok {
		dialOpts = append(dialOpts, libmqtt.WithDialTimeout(uint16(time.Until(dd).Seconds())))
	}

	err := c.client.ConnectServer(c.brokerAddress, dialOpts...)
	if err != nil {
		return err
	}
	select {
	case <-dialCtx.Done():
		return dialCtx.Err()
	case err = <-c.netErrCh:
		if err != nil {
			return err
		}
	case err = <-c.connErrCh:
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Start(ctx context.Context) error {
	c.client.HandleTopic(c.cmdSubTopicHandle, c.handleTopicMsg)
	c.client.Subscribe(&libmqtt.Topic{Name: c.cmdSubTopic, Qos: libmqtt.Qos1})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c.subErrCh:
		if err != nil {
			return err
		}
	}

	// publish a packet to notify aranya we are online
	c.Log.D("publishing online message")
	c.pubOnline()

	select {
	case err := <-c.netErrCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) PostMsg(msg *aranyagopb.Msg) error {
	data, err := msg.Marshal()
	if err != nil {
		return err
	}

	c.client.Publish(&libmqtt.PublishPacket{TopicName: c.pubTopic, Qos: libmqtt.Qos1, Payload: data})

	return nil
}

func (c *Client) Close() error {
	return c.OnClose(func() error {
		c.client.Destroy(true)

		if atomic.CompareAndSwapInt32(&c.exited, 0, 1) {
			close(c.netErrCh)
		}

		return nil
	})
}

func (c *Client) pubOnline() {
	c.client.Publish(&libmqtt.PublishPacket{
		TopicName: c.pubWillTopic,
		Qos:       libmqtt.Qos1,
		Payload:   c.onlineMsg,
		IsRetain:  c.supportRetain,
	})
}

func (c *Client) handleNet(client libmqtt.Client, server string, err error) {
	if err != nil {
		c.Log.I("network error happened", log.String("server", server), log.Error(err))

		// exit client on network error
		if atomic.CompareAndSwapInt32(&c.exited, 0, 1) {
			c.netErrCh <- err
			close(c.netErrCh)
		}
	}
}

// nolint:gocritic
func (c *Client) handleConn(dialExitSig <-chan struct{}) libmqtt.ConnHandleFunc {
	return func(client libmqtt.Client, server string, code byte, err error) {
		if err != nil {
			select {
			case <-dialExitSig:
				return
			case c.connErrCh <- err:
				return
			}
		} else if code != libmqtt.CodeSuccess {
			select {
			case <-dialExitSig:
				return
			case c.connErrCh <- fmt.Errorf("rejected by mqtt broker, code: %d", code):
				return
			}
		} else {
			// connected to broker
			select {
			case <-dialExitSig:
				return
			case c.connErrCh <- nil:
				return
			}
		}
	}
}

func (c *Client) handleSub(dialExitSig <-chan struct{}) libmqtt.SubHandleFunc {
	return func(client libmqtt.Client, topics []*libmqtt.Topic, err error) {
		select {
		case <-dialExitSig:
			return
		case c.subErrCh <- err:
			return
		}
	}
}

func (c *Client) handlePub(client libmqtt.Client, topic string, err error) {
	if err != nil {
		c.Log.I("failed to publish message", log.String("topic", topic), log.Error(err))
		if topic == c.pubWillTopic {
			c.Log.D("republishing online message")
			c.pubOnline()
		}
	}
}

func (c *Client) handleTopicMsg(client libmqtt.Client, topic string, qos libmqtt.QosLevel, cmdBytes []byte) {
	cmd := new(aranyagopb.Cmd)
	err := cmd.Unmarshal(cmdBytes)
	if err != nil {
		c.Log.I("failed to unmarshal cmd", log.Binary("cmdBytes", cmdBytes), log.Error(err))
		return
	}

	c.HandleCmd(cmd)
}
