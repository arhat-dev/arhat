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
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"
	"github.com/goiiot/libmqtt"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
)

func NewMQTTClient(agent types.Agent, config *conf.ArhatMQTTConfig) (_ types.AgentConnectivity, err error) {
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

	onlineMsgBytes, _ := aranyagopb.NewOnlineMsg(connInfo.ClientID).Marshal()
	willMsgBytes, _ := aranyagopb.NewOfflineMsg(connInfo.ClientID).Marshal()

	if connInfo.TLSConfig != nil {
		options = append(options, libmqtt.WithCustomTLS(connInfo.TLSConfig))
	}

	keepalive := config.Keepalive
	if keepalive == 0 {
		// default to 60 seconds
		keepalive = 60
	}

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

	c := &MQTTClient{
		baseClient:    newBaseClient(agent, connInfo.MaxDataSize),
		client:        client,
		supportRetain: connInfo.SupportRetain,

		brokerAddress:     config.Endpoint,
		cmdSubTopicHandle: connInfo.CmdSubTopicHandle,
		cmdSubTopic:       connInfo.CmdSubTopic,
		pubTopic:          connInfo.MsgPubTopic,
		pubWillTopic:      connInfo.WillPubTopic,
		onlineMsg:         onlineMsgBytes,
		clientID:          connInfo.ClientID,

		netErrCh:  make(chan error),
		connErrCh: make(chan error),
		subErrCh:  make(chan error),
	}

	return c, nil
}

type MQTTClient struct {
	onlineMsg         []byte
	brokerAddress     string
	cmdSubTopicHandle string
	cmdSubTopic       string
	pubTopic          string
	pubWillTopic      string
	clientID          string

	*baseClient
	client libmqtt.Client

	netErrCh  chan error
	connErrCh chan error
	subErrCh  chan error

	exited        int32
	supportRetain bool
}

func (c *MQTTClient) Connect(dialCtx context.Context) error {
	dialOpts := []libmqtt.Option{
		libmqtt.WithRouter(libmqtt.NewRegexRouter()),
		libmqtt.WithAutoReconnect(false),
		libmqtt.WithConnHandleFunc(c.handleConn(c.ctx.Done())),
		libmqtt.WithSubHandleFunc(c.handleSub(c.ctx.Done())),
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

func (c *MQTTClient) Start(ctx context.Context) error {
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
	c.log.D("publishing online message")
	c.pubOnline()

	select {
	case err := <-c.netErrCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *MQTTClient) PostMsg(msg *aranyagopb.Msg) error {
	msg.OnlineId = c.clientID

	data, err := msg.Marshal()
	if err != nil {
		return err
	}

	c.client.Publish(&libmqtt.PublishPacket{TopicName: c.pubTopic, Qos: libmqtt.Qos1, Payload: data})

	return nil
}

func (c *MQTTClient) Close() error {
	c.client.Destroy(true)

	if atomic.CompareAndSwapInt32(&c.exited, 0, 1) {
		close(c.netErrCh)
	}

	c.exit()
	return nil
}

func (c *MQTTClient) pubOnline() {
	c.client.Publish(&libmqtt.PublishPacket{
		TopicName: c.pubWillTopic,
		Qos:       libmqtt.Qos1,
		Payload:   c.onlineMsg,
		IsRetain:  c.supportRetain,
	})
}

func (c *MQTTClient) handleNet(client libmqtt.Client, server string, err error) {
	if err != nil {
		c.log.I("network error happened", log.String("server", server), log.Error(err))

		// exit client on network error
		if atomic.CompareAndSwapInt32(&c.exited, 0, 1) {
			c.netErrCh <- err
			close(c.netErrCh)
		}
	}
}

// nolint:gocritic
func (c *MQTTClient) handleConn(dialExitSig <-chan struct{}) libmqtt.ConnHandleFunc {
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

func (c *MQTTClient) handleSub(dialExitSig <-chan struct{}) libmqtt.SubHandleFunc {
	return func(client libmqtt.Client, topics []*libmqtt.Topic, err error) {
		select {
		case <-dialExitSig:
			return
		case c.subErrCh <- err:
			return
		}
	}
}

func (c *MQTTClient) handlePub(client libmqtt.Client, topic string, err error) {
	if err != nil {
		c.log.I("failed to publish message", log.String("topic", topic), log.Error(err))
		if topic == c.pubWillTopic {
			c.log.D("republishing online message")
			c.pubOnline()
		}
	}
}

func (c *MQTTClient) handleTopicMsg(client libmqtt.Client, topic string, qos libmqtt.QosLevel, cmdBytes []byte) {
	cmd := new(aranyagopb.Cmd)
	err := cmd.Unmarshal(cmdBytes)
	if err != nil {
		c.log.I("failed to unmarshal cmd", log.Binary("cmdBytes", cmdBytes), log.Error(err))
		return
	}

	c.parent.HandleCmd(cmd)
}
