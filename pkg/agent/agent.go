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

package agent

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"
	"sync/atomic"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/libext/server"
	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"
	"ext.arhat.dev/runtimeutil/networkutil"
	"ext.arhat.dev/runtimeutil/storageutil"
	"github.com/gogo/protobuf/proto"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
	"arhat.dev/arhat/pkg/util/errconv"
	"arhat.dev/arhat/pkg/util/manager"
)

var (
	errClientNotSet            = errors.New("client not set")
	errRequiredOptionsNotFound = errors.New("required options not found")
	errStreamSessionClosed     = errors.New("stream session closed")
	errCommandNotProvided      = errors.New("command not provided for exec")
)

var _ types.Agent = (*Agent)(nil)

type (
	rawCmdHandleFunc func(sid uint64, data []byte)
)

func NewAgent(appCtx context.Context, logger log.Interface, config *conf.Config) (_ *Agent, err error) {
	extInfo, err := convertNodeExtInfo(config.Arhat.Node.ExtInfo)
	if err != nil {
		return nil, err
	}

	sc, err := config.Storage.CreateClient(appCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	nc := networkutil.NewClient(
		func(
			ctx context.Context,
			env map[string]string,
			stdin io.Reader,
			stdout, stderr io.Writer,
		) error {
			if !config.Network.Enabled {
				return wellknownerrors.ErrNotSupported
			}

			// execute host command for abbot request
			cmd, err := exechelper.Do(exechelper.Spec{
				Context: ctx,
				Command: config.Network.AbbotRequestExec,
				Env:     env,
				Stdin:   stdin,
				Stdout:  stdout,
				Stderr:  stderr,
			})
			if err != nil {
				return err
			}

			_, err = cmd.Wait()
			return err
		},
	)

	agent := &Agent{
		hostConfig:    &config.Arhat.Host,
		machineIDFrom: &config.Arhat.Node.MachineIDFrom,
		kubeLogFile:   config.Arhat.Log.KubeLogFile(),
		extInfo:       extInfo,

		ctx: appCtx,

		logger: logger,

		metricsMU: new(sync.RWMutex),

		cmdMgr:        manager.NewCmdManager(),
		storageClient: sc,
		networkClient: nc,

		streams: manager.NewStreamManager(),

		gzipPool: &sync.Pool{
			New: func() interface{} {
				w, _ := gzip.NewWriterLevel(nil, gzip.BestCompression)
				return w
			},
		},
	}

	agent.funcMap = map[aranyagopb.CmdType]rawCmdHandleFunc{
		aranyagopb.CMD_SESSION_CLOSE: agent.handleSessionClose,
		aranyagopb.CMD_REJECT:        agent.handleRejectCmd,

		aranyagopb.CMD_NET:     agent.handleNetwork,
		aranyagopb.CMD_RUNTIME: agent.handleRuntime,

		aranyagopb.CMD_NODE_INFO_GET: agent.handleNodeInfoGet,
		aranyagopb.CMD_EXEC:          agent.handlePodContainerExec,
		aranyagopb.CMD_ATTACH:        agent.handlePodContainerAttach,
		aranyagopb.CMD_LOGS:          agent.handlePodContainerLogs,
		aranyagopb.CMD_TTY_RESIZE:    agent.handlePodContainerTerminalResize,
		aranyagopb.CMD_PORT_FORWARD:  agent.handlePodPortForward,

		aranyagopb.CMD_METRICS_CONFIG:  agent.handleMetricsConfig,
		aranyagopb.CMD_METRICS_COLLECT: agent.handleMetricsCollect,

		aranyagopb.CMD_CRED_ENSURE: agent.handleCredentialEnsure,

		aranyagopb.CMD_PERIPHERAL_LIST:            agent.handlePeripheralList,
		aranyagopb.CMD_PERIPHERAL_ENSURE:          agent.handlePeripheralEnsure,
		aranyagopb.CMD_PERIPHERAL_DELETE:          agent.handlePeripheralDelete,
		aranyagopb.CMD_PERIPHERAL_OPERATE:         agent.handlePeripheralOperation,
		aranyagopb.CMD_PERIPHERAL_COLLECT_METRICS: agent.handlePeripheralMetricsCollect,
	}

	if config.Extension.Enabled {
		var endpoints []server.EndpointConfig
		for _, ep := range config.Extension.Endpoints {
			tlsConfig, err := ep.TLS.TLSConfig.GetTLSConfig(true)
			if err != nil {
				return nil, fmt.Errorf("failed to create tls config for extension endpoint %q: %w", ep.Listen, err)
			}

			if tlsConfig != nil && ep.TLS.VerifyClientCert {
				tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			}

			endpoints = append(endpoints, server.EndpointConfig{
				Listen:            ep.Listen,
				TLS:               tlsConfig,
				KeepaliveInterval: ep.KeepaliveInterval,
				MessageTimeout:    ep.MessageTimeout,
			})
		}
		srv, err := server.NewServer(agent.ctx, agent.logger, &server.Config{
			Endpoints: endpoints,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create extension server")
		}

		agent.createAndRegisterPeripheralExtensionManager(agent.ctx, srv, &config.Extension.Peripherals)

		go func() {
			err2 := srv.ListenAndServe()
			if err2 != nil {
				panic(err2)
			}
		}()
	}

	agent.streams.Start(appCtx.Done())

	return agent, nil
}

type Agent struct {
	ctx context.Context

	hostConfig    *conf.HostConfig
	machineIDFrom *conf.ValueFromSpec
	kubeLogFile   string
	extInfo       []*aranyagopb.NodeExtInfo

	logger log.Interface

	metricsMU               *sync.RWMutex
	collectNodeMetrics      types.MetricsCollectFunc
	collectContainerMetrics types.MetricsCollectFunc

	cmdMgr *manager.CmdManager

	storageClient *storageutil.Client
	networkClient *networkutil.Client

	streams *manager.StreamManager

	agentComponentPeripheral

	gzipPool *sync.Pool

	settingClient uint32
	client        types.ConnectivityClient

	funcMap map[aranyagopb.CmdType]rawCmdHandleFunc
}

func (b *Agent) SetClient(client types.ConnectivityClient) {
	for !atomic.CompareAndSwapUint32(&b.settingClient, 0, 1) {
		runtime.Gosched()
	}

	b.client = client

	for !atomic.CompareAndSwapUint32(&b.settingClient, 1, 0) {
		runtime.Gosched()
	}
}

func (b *Agent) GetClient() types.ConnectivityClient {
	for !atomic.CompareAndSwapUint32(&b.settingClient, 0, 1) {
		runtime.Gosched()
	}

	defer func() {
		for !atomic.CompareAndSwapUint32(&b.settingClient, 1, 0) {
			runtime.Gosched()
		}
	}()

	return b.client
}

func (b *Agent) GetGzipWriter(w io.Writer) *gzip.Writer {
	gw := b.gzipPool.Get().(*gzip.Writer)
	gw.Reset(w)
	return gw
}

func (b *Agent) PostData(sid uint64, kind aranyagopb.MsgType, seq uint64, completed bool, data []byte) (uint64, error) {
	client := b.GetClient()
	if client == nil {
		return seq, errClientNotSet
	}

	n := client.MaxPayloadSize()
	for len(data) > n {
		buf := data
		err := client.PostMsg(aranyagopb.NewMsg(kind, sid, seq, false, buf[:n]))
		if err != nil {
			return seq, fmt.Errorf("failed to post msg chunk: %w", err)
		}
		seq++
		data = data[n:]
	}

	err := client.PostMsg(aranyagopb.NewMsg(kind, sid, seq, completed, data))
	if err != nil {
		return seq, fmt.Errorf("failed to post msg chunk: %w", err)
	}
	return seq, nil
}

func (b *Agent) PostMsg(sid uint64, kind aranyagopb.MsgType, msg proto.Marshaler) error {
	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal msg body: %w", err)
	}

	_, err = b.PostData(sid, kind, 0, true, data)
	return err
}

func (b *Agent) HandleCmd(cmd *aranyagopb.Cmd) {
	if b.GetClient() == nil {
		b.handleConnectivityError(0, fmt.Errorf("client empty"))
		return
	}

	sid := cmd.Sid

	// handle stream data first
	if cmd.Kind == aranyagopb.CMD_DATA_UPSTREAM {
		if cmd.Completed {
			b.streams.CloseRead(sid, cmd.Seq)
			return
		}

		if !b.streams.Write(sid, cmd.Seq, cmd.Payload) {
			b.handleRuntimeError(sid, errStreamSessionClosed)
			return
		}

		return
	}

	cmdBytes, complete := b.cmdMgr.Process(cmd)
	if !complete {
		return
	}

	handleCmd, ok := b.funcMap[cmd.Kind]
	if handleCmd == nil || !ok {
		b.handleUnknownCmd(sid, "unknown or unsupported cmd", cmd)
		return
	}

	handleCmd(sid, cmdBytes)
}

func (b *Agent) processInNewGoroutine(sid uint64, cmdName string, process func()) {
	go func() {
		b.logger.V("working on", log.Uint64("sid", sid), log.String("work", cmdName))
		process()
		b.logger.V("finished", log.Uint64("sid", sid), log.String("work", cmdName))
	}()
}

// nolint:unparam
func (b *Agent) handleUnknownCmd(sid uint64, category string, cmd interface{}) bool {
	b.logger.I(fmt.Sprintf("unknown %s cmd", category), log.Uint64("sid", sid), log.Any("cmd", cmd))
	return b.handleRuntimeError(sid, wellknownerrors.ErrNotSupported)
}

func (b *Agent) handleRuntimeError(sid uint64, err error) bool {
	if err == nil {
		return false
	}

	plainErr := b.PostMsg(sid, aranyagopb.MSG_ERROR, errconv.ToConnectivityError(err))
	if plainErr != nil {
		b.handleConnectivityError(sid, plainErr)
	}

	return true
}

// nolint:unparam
func (b *Agent) handleConnectivityError(sid uint64, err error) bool {
	if err == nil {
		return false
	}

	b.logger.I("connectivity error", log.Uint64("sid", sid), log.Error(err))
	return true
}
