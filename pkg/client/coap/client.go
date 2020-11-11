package coap

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/log"
	piondtls "github.com/pion/dtls/v2"
	"github.com/pion/logging"
	coapdtls "github.com/plgd-dev/go-coap/v2/dtls"
	coapmsg "github.com/plgd-dev/go-coap/v2/message"
	coapmsgcodes "github.com/plgd-dev/go-coap/v2/message/codes"
	coapmux "github.com/plgd-dev/go-coap/v2/mux"
	coapnet "github.com/plgd-dev/go-coap/v2/net"
	"github.com/plgd-dev/go-coap/v2/net/keepalive"
	coaptcp "github.com/plgd-dev/go-coap/v2/tcp"
	coapudp "github.com/plgd-dev/go-coap/v2/udp"
	coapudpclient "github.com/plgd-dev/go-coap/v2/udp/client"
	coapudpmsgpool "github.com/plgd-dev/go-coap/v2/udp/message/pool"

	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/types"
)

// nolint:gocyclo
func NewCoAPClient(
	ctx context.Context,
	handleCmd types.AgentCmdHandleFunc,
	cfg interface{},
) (_ types.ConnectivityClient, err error) {
	config, ok := cfg.(*ConnectivityCoAP)
	if !ok {
		return nil, fmt.Errorf("unepxected non coap config")
	}

	var (
		tlsCfg         *tls.Config
		connectActions []func(dialCtx context.Context) (*coapudpclient.ClientConn, error)
	)

	if config.TLS.Enabled {
		tlsCfg, err = config.TLS.GetTLSConfig(false)
		if err != nil {
			return nil, err
		}
	}

	pathNamespace, err := config.PathNamespaceFrom.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get path namespace value: %w", err)
	}

	observeTopic, putTopic, willTopic := aranyagoconst.CoAPTopics(pathNamespace)

	willMsgOpts, _, err := new(coapmsg.Options).SetPath(make([]byte, len(willTopic)*2), willTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to create will msg opts")
	}

	putMsgOpts, _, err := new(coapmsg.Options).SetPath(make([]byte, len(putTopic)*2), putTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to create will msg opts")
	}

	obMsgOpts, _, err := new(coapmsg.Options).SetPath(make([]byte, len(observeTopic)*2), observeTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to create observe msg opts")
	}

	allOpts := []coapmsg.Options{willMsgOpts, putMsgOpts, obMsgOpts}
	for i := range allOpts {
		for k, v := range config.URIQueries {
			s := k + "=" + v

			allOpts[i], _, err = allOpts[i].AddString(make([]byte, len(s)*2), coapmsg.URIQuery, s)
			if err != nil {
				return nil, fmt.Errorf("failed to add uri query: %w", err)
			}
		}
	}
	willMsgOpts, putMsgOpts, obMsgOpts = allOpts[0], allOpts[1], allOpts[2]

	maxPayloadSize := config.MaxPayloadSize
	if maxPayloadSize <= 0 {
		maxPayloadSize = aranyagoconst.MaxCoAPDataSize
	}

	coapClient := &Client{
		willMsgOpts: willMsgOpts,
		putMsgOpts:  putMsgOpts,
		obMsgOpts:   obMsgOpts,

		mu: new(sync.Mutex),
	}

	coapClient.BaseClient, err = clientutil.NewBaseClient(ctx, handleCmd, maxPayloadSize)
	if err != nil {
		return nil, err
	}

	transport := config.Transport
	if transport == "" {
		transport = "udp"
	}

	brokerAddr := config.Endpoint

	keepaliveInterval := time.Second * time.Duration(config.Keepalive)
	if keepaliveInterval == 0 {
		keepaliveInterval = time.Second * 6
	}
	keepaliveOpt := keepalive.New(keepalive.WithConfig(keepalive.Config{
		Interval:    keepaliveInterval,
		WaitForPong: keepaliveInterval,
		NewRetryPolicy: func() keepalive.RetryFunc {
			// The first failure is detected after 2*duration:
			// 1 since the previous ping, plus 1 for the next ping-pong to timeout
			start := time.Now()
			attempt := time.Duration(1)
			return func() (time.Time, error) {
				attempt++
				// Try to send ping and wait for pong 2 more times
				if time.Since(start) <= 2*2*keepaliveInterval {
					return start.Add(attempt * keepaliveInterval), nil
				}
				return time.Time{}, keepalive.ErrKeepAliveDeadlineExceeded
			}
		},
	}))

	// no server provided, run in multicast mode
	if brokerAddr == "" {
		srvOpts := []coapudp.ServerOption{
			coapudp.WithMaxMessageSize(aranyagoconst.MaxCoAPDataSize),
		}

		connectActions = append(connectActions, func(dialCtx context.Context) (*coapudpclient.ClientConn, error) {
			udpListener, err2 := coapnet.NewListenUDP(config.Transport, ":0")
			if err2 != nil {
				return nil, fmt.Errorf("failed to listen random udp port for mcast coap: %w", err2)
			}

			srv := coapudp.NewServer(srvOpts...)
			defer func() {
				srv.Stop()
				_ = udpListener.Close()
			}()

			go func() {
				_ = srv.Serve(udpListener)
			}()

			return nil, srv.Discover(dialCtx,
				"224.0.1.187:5683", "/oic/res",
				func(cc *coapudpclient.ClientConn, resp *coapudpmsgpool.Message) {
					// nolint:gocritic
					switch resp.Code() {
					case coapmsgcodes.CSM:
						// TODO
					}
					brokerAddr = cc.RemoteAddr().String()
				},
			)
		})
	}

	switch strings.ToLower(transport) {
	case "tcp", "tcp4", "tcp6":
		dialOpts := []coaptcp.DialOption{
			coaptcp.WithMaxMessageSize(aranyagoconst.MaxCoAPDataSize),
			coaptcp.WithKeepAlive(keepaliveOpt),
			coaptcp.WithErrors(func(err error) {
				coapClient.Log.I("internal coap client error", log.Error(err))
			}),
		}

		if tlsCfg != nil {
			dialOpts = append(dialOpts, coaptcp.WithTLS(tlsCfg))
		}

		coapClient.connect = func(dialCtx context.Context) (coapmux.Client, error) {
			dialOpts = []coaptcp.DialOption{
				coaptcp.WithContext(dialCtx),
			}

			conn, err2 := coaptcp.Dial(brokerAddr, dialOpts...)
			if err2 != nil {
				return nil, fmt.Errorf("failed to dial tcp server: %w", err2)
			}

			return coaptcp.NewClientTCP(conn), nil
		}
	case "udp", "udp4", "udp6":
		if tlsCfg == nil {
			connectActions = append(connectActions, func(dialCtx context.Context) (*coapudpclient.ClientConn, error) {
				dialOpts := []coapudp.DialOption{
					coapudp.WithMaxMessageSize(aranyagoconst.MaxCoAPDataSize),
					coapudp.WithContext(dialCtx),
					coapudp.WithKeepAlive(keepaliveOpt),
					coapudp.WithErrors(func(err error) {
						coapClient.Log.I("internal coap client error", log.Error(err))
					}),
				}

				conn, err2 := coapudp.Dial(brokerAddr, dialOpts...)
				if err2 != nil {
					return nil, fmt.Errorf("failed to dial udp server: %w", err2)
				}

				return conn, nil
			})
		} else {
			dtlsConfig := &piondtls.Config{
				Certificates:       tlsCfg.Certificates,
				ClientAuth:         piondtls.ClientAuthType(tlsCfg.ClientAuth),
				InsecureSkipVerify: tlsCfg.InsecureSkipVerify,
				RootCAs:            tlsCfg.RootCAs,
				ClientCAs:          tlsCfg.ClientCAs,
				InsecureHashes:     config.TLS.AllowInsecureHashes,
				ServerName:         tlsCfg.ServerName,
				LoggerFactory:      &loggerFactory{coapClient.Log.WithName("dtls")},
			}

			var (
				defaultSecret []byte
				mapping       = make(map[string][]byte)
			)
			for i, kv := range config.TLS.PreSharedKey.ServerHintMapping {
				parts := strings.SplitN(kv, ":", 2)
				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid server hint mapping at index %d", i)
				}

				k, v := parts[0], parts[1]
				secret, err2 := base64.StdEncoding.DecodeString(v)
				if err2 != nil {
					return nil, fmt.Errorf("failed to decode secret at index %d: %w", i, err2)
				}

				if k == "" {
					defaultSecret = secret
				} else {
					srvHint, err3 := base64.StdEncoding.DecodeString(k)
					if err3 != nil {
						return nil, fmt.Errorf("failed to decode secret at index %d: %w", i, err3)
					}

					mapping[string(srvHint)] = secret
				}
			}

			if len(defaultSecret) != 0 || len(mapping) != 0 {
				dtlsConfig.PSK = func(hint []byte) ([]byte, error) {
					if k, ok := mapping[string(hint)]; ok {
						return k, nil
					}

					if len(defaultSecret) != 0 {
						return defaultSecret, nil
					}

					return nil, fmt.Errorf("no server hint matched")
				}
			}

			if idHint := config.TLS.PreSharedKey.IdentityHint; idHint != "" {
				dtlsConfig.PSKIdentityHint, err = base64.StdEncoding.DecodeString(idHint)
				if err != nil {
					return nil, fmt.Errorf("invalid psk identity hint: %w", err)
				}
			}

			for _, c := range tlsCfg.CipherSuites {
				cid := piondtls.CipherSuiteID(c)
				if cid.String() == fmt.Sprintf("unknown(%v)", c) {
					return nil, fmt.Errorf("unsupported cipher suites")
				}

				dtlsConfig.CipherSuites = append(dtlsConfig.CipherSuites, cid)
			}

			connectActions = append(connectActions, func(dialCtx context.Context) (*coapudpclient.ClientConn, error) {
				dialOpts := []coapdtls.DialOption{
					coapdtls.WithMaxMessageSize(aranyagoconst.MaxCoAPDataSize),
					coapdtls.WithContext(dialCtx),
					coapdtls.WithKeepAlive(keepaliveOpt),
					coapdtls.WithErrors(func(err error) {
						coapClient.Log.I("internal coap client error", log.Error(err))
					}),
				}

				dtlsConn, err := coapdtls.Dial(brokerAddr, dtlsConfig, dialOpts...)
				if err != nil {
					return nil, fmt.Errorf("failed to dial dtls server: %w", err)
				}

				return dtlsConn, nil
			})

		}

		coapClient.connect = func(dialCtx context.Context) (coapmux.Client, error) {
			for _, doConnect := range connectActions {
				cc, err := doConnect(dialCtx)
				if err != nil {
					return nil, err
				}

				if cc != nil {
					// this is the last connect action
					return coapudpclient.NewClient(cc), nil
				}
			}

			// unreachable
			return nil, fmt.Errorf("unreachable code")
		}
	}

	return coapClient, nil
}

type Client struct {
	*clientutil.BaseClient

	connected, started uint32

	connect func(dialCtx context.Context) (coapmux.Client, error)
	client  coapmux.Client

	willMsgOpts coapmsg.Options
	putMsgOpts  coapmsg.Options
	obMsgOpts   coapmsg.Options

	clientID          string
	cancelObservation func(ctx context.Context) error

	mu *sync.Mutex
}

func (c *Client) Connect(dialCtx context.Context) error {
	if !atomic.CompareAndSwapUint32(&c.connected, 0, 1) {
		return clientutil.ErrClientAlreadyConnected
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	c.client, err = c.connect(dialCtx)

	return err
}

func (c *Client) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapUint32(&c.started, 0, 1) {
		return clientutil.ErrClientNotConnected
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	obs, err := c.client.Observe(ctx, "", c.handleCmdRecv, c.obMsgOpts...)
	if err != nil {
		return fmt.Errorf("failed to observe cmd topic: %w", err)
	}

	payload, _ := aranyagopb.NewOnlineStateMsg(c.clientID).Marshal()
	msg := aranyagopb.NewMsg(aranyagopb.MSG_STATE, 0, 0, true, payload)
	err = c.doPostMsg(c.Context(), msg, c.willMsgOpts)
	if err != nil {
		return fmt.Errorf("failed to publish online msg ")
	}
	c.cancelObservation = obs.Cancel

	select {
	case <-ctx.Done():
		return nil
	case <-c.Context().Done():
		return nil
	}
}

func (c *Client) doPostMsg(ctx context.Context, msg *aranyagopb.Msg, msgOpts coapmsg.Options) error {
	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal msg: %w", err)
	}

	coapMsg := &coapmsg.Message{
		Context: ctx,
		Code:    coapmsgcodes.PUT,
		Options: msgOpts,
		Body:    bytes.NewReader(data),
	}

	coapMsg.Token, err = coapmsg.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get random token: %w", err)
	}

	resp, err := c.client.Do(coapMsg)
	if err != nil {
		return fmt.Errorf("failed to publish msg: %w", err)
	}

	switch resp.Code {
	case coapmsgcodes.Created, coapmsgcodes.Changed:
		// publish successful
	default:
		return fmt.Errorf("failed to publish msg: %s", resp.Code.String())
	}

	return nil
}

func (c *Client) PostMsg(msg *aranyagopb.Msg) error {
	return c.doPostMsg(c.Context(), msg, c.putMsgOpts)
}

func (c *Client) Close() error {
	return c.OnClose(func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.cancelObservation != nil {
			_ = c.cancelObservation(context.TODO())
		}

		if c.client != nil {
			// TODO: currently we only send offline message with best effort
			//		 need to ensure the offline message is acknowledged by aranya
			payload, _ := aranyagopb.NewOfflineStateMsg(c.clientID).Marshal()
			msg := aranyagopb.NewMsg(aranyagopb.MSG_STATE, 0, 0, true, payload)
			_ = c.doPostMsg(context.Background(), msg, c.willMsgOpts)

			return c.client.Close()
		}

		return nil
	})
}

func (c *Client) handleCmdRecv(notification *coapmsg.Message) {
	if notification == nil || notification.Body == nil {
		return
	}

	if notification.Code != coapmsgcodes.Content {
		c.Log.I("unexpected cmd data", log.StringError(notification.Code.String()))
		return
	}

	cmdData, err := ioutil.ReadAll(notification.Body)
	if err != nil {
		c.Log.I("failed to read cmd data", log.Error(err))
		return
	}

	cmd := new(aranyagopb.Cmd)
	err = cmd.Unmarshal(cmdData)
	if err != nil {
		c.Log.I("failed to unmarshal cmd data", log.Error(err))
	} else {
		c.HandleCmd(cmd)
	}
}

type loggerFactory struct {
	log.Interface
}

func (l *loggerFactory) NewLogger(scope string) logging.LeveledLogger {
	return &leveledLogger{l.Interface.WithName(scope)}
}

type leveledLogger struct {
	log.Interface
}

func (l *leveledLogger) Trace(msg string) {
	l.V(msg)
}

func (l *leveledLogger) Tracef(format string, args ...interface{}) {
	l.V(fmt.Sprintf(format, args...))
}

func (l *leveledLogger) Debug(msg string) {
	l.D(msg)
}

func (l *leveledLogger) Debugf(format string, args ...interface{}) {
	l.D(fmt.Sprintf(format, args...))
}

func (l *leveledLogger) Info(msg string) {
	l.I(msg)
}

func (l *leveledLogger) Infof(format string, args ...interface{}) {
	l.I(fmt.Sprintf(format, args...))
}

func (l *leveledLogger) Warn(msg string) {
	l.I(msg)
}

func (l *leveledLogger) Warnf(format string, args ...interface{}) {
	l.I(fmt.Sprintf(format, args...))
}

func (l *leveledLogger) Error(msg string) {
	l.E(msg)
}

func (l *leveledLogger) Errorf(format string, args ...interface{}) {
	l.E(fmt.Sprintf(format, args...))
}
