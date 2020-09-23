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
	"context"
	"errors"
	"fmt"
	goruntime "runtime"
	"sync"
	"sync/atomic"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"
	"github.com/gogo/protobuf/proto"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/runtime"
	"arhat.dev/arhat/pkg/runtime/none"
	"arhat.dev/arhat/pkg/storage"
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

var _ types.Agent = &Agent{}

func NewAgent(appCtx context.Context, config *conf.ArhatConfig) (*Agent, error) {
	ctx, exit := context.WithCancel(appCtx)

	extInfo, err := convertNodeExtInfo(config.Arhat.Node.ExtInfo)

	defer func() {
		if err != nil {
			exit()
		}
	}()

	if err != nil {
		return nil, err
	}

	st, err := storage.NewStorage(ctx, &config.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage handler: %w", err)
	}

	var rt types.Runtime
	if config.Runtime.Enabled {
		rt, err = runtime.NewRuntime(appCtx, st, &config.Runtime)
	} else {
		rt, err = none.NewNoneRuntime(appCtx, st, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	rt.SetContext(ctx)

	err = rt.InitRuntime()
	if err != nil {
		return nil, fmt.Errorf("failed to init runtime: %w", err)
	}

	agent := &Agent{
		hostConfig:    &config.Arhat.Host,
		machineIDFrom: &config.Arhat.Node.MachineIDFrom,
		kubeLogFile:   config.Arhat.Log.KubeLogFile(),
		extInfo:       extInfo,

		ctx:  ctx,
		exit: exit,

		logger: log.Log.WithName("agent"),

		metricsMU: new(sync.RWMutex),

		cmdMgr:  manager.NewCmdManager(),
		runtime: rt,
		storage: st,
		devices: newDeviceManager(),
		streams: manager.NewStreamManager(),
	}
	agent.streams.Start(ctx.Done())

	return agent, nil
}

type Agent struct {
	hostConfig    *conf.HostConfig
	machineIDFrom *conf.ValueFromSpec
	kubeLogFile   string
	extInfo       []*aranyagopb.NodeExtInfo

	ctx  context.Context
	exit context.CancelFunc

	logger log.Interface

	metricsMU               *sync.RWMutex
	collectNodeMetrics      types.MetricsCollectFunc
	collectContainerMetrics types.MetricsCollectFunc

	cmdMgr  *manager.CmdManager
	runtime types.Runtime
	storage types.Storage
	devices *deviceManager
	streams *manager.StreamManager

	settingClient uint32
	client        types.ConnectivityClient
}

func (b *Agent) SetClient(client types.ConnectivityClient) {
	for !atomic.CompareAndSwapUint32(&b.settingClient, 0, 1) {
		goruntime.Gosched()
	}

	b.client = client

	for !atomic.CompareAndSwapUint32(&b.settingClient, 1, 0) {
		goruntime.Gosched()
	}
}

func (b *Agent) GetClient() types.ConnectivityClient {
	for !atomic.CompareAndSwapUint32(&b.settingClient, 0, 1) {
		goruntime.Gosched()
	}

	defer func() {
		for !atomic.CompareAndSwapUint32(&b.settingClient, 1, 0) {
			goruntime.Gosched()
		}
	}()

	return b.client
}

func (b *Agent) PostData(sid uint64, kind aranyagopb.Kind, seq uint64, completed bool, data []byte) (uint64, error) {
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

func (b *Agent) PostMsg(sid uint64, kind aranyagopb.Kind, msg proto.Marshaler) error {
	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal msg body: %w", err)
	}

	_, err = b.PostData(sid, kind, 0, true, data)
	return err
}

func (b *Agent) Context() context.Context {
	return b.ctx
}

func (b *Agent) HandleCmd(cmd *aranyagopb.Cmd) {
	if b.GetClient() == nil {
		b.handleConnectivityError(0, fmt.Errorf("client empty"))
		return
	}

	sid := cmd.Header.Sid

	// handle stream data first
	if cmd.Header.Kind == aranyagopb.CMD_DATA_UPSTREAM {
		if cmd.Header.Completed {
			b.streams.CloseRead(sid, cmd.Header.Seq)
			return
		}

		if !b.streams.Write(sid, cmd.Header.Seq, cmd.Body) {
			b.handleRuntimeError(sid, errStreamSessionClosed)
			return
		}

		return
	}

	cmdBytes, complete := b.cmdMgr.Process(cmd)
	if !complete {
		return
	}

	handleCmd, ok := map[aranyagopb.Kind]types.RawCmdHandleFunc{
		aranyagopb.CMD_SESSION_CLOSE: b.handleSessionClose,

		aranyagopb.CMD_NODE_INFO_GET: b.handleNodeInfoGet,

		aranyagopb.CMD_DEVICE_LIST:   b.handleDeviceList,
		aranyagopb.CMD_DEVICE_ENSURE: b.handleDeviceEnsure,
		aranyagopb.CMD_DEVICE_DELETE: b.handleDeviceDelete,

		aranyagopb.CMD_METRICS_CONFIG:  b.handleMetricsConfig,
		aranyagopb.CMD_METRICS_COLLECT: b.handleMetricsCollect,

		aranyagopb.CMD_IMAGE_LIST:   nil,
		aranyagopb.CMD_IMAGE_ENSURE: b.handleImageEnsure,
		aranyagopb.CMD_IMAGE_DELETE: nil,

		aranyagopb.CMD_POD_LIST:   b.handlePodList,
		aranyagopb.CMD_POD_ENSURE: b.handlePodEnsure,
		aranyagopb.CMD_POD_DELETE: b.handlePodDelete,

		aranyagopb.CMD_POD_CTR_EXEC:       b.handlePodContainerExec,
		aranyagopb.CMD_POD_CTR_ATTACH:     b.handlePodContainerAttach,
		aranyagopb.CMD_POD_CTR_LOGS:       b.handlePodContainerLogs,
		aranyagopb.CMD_POD_CTR_TTY_RESIZE: b.handlePodContainerTerminalResize,
		aranyagopb.CMD_POD_PORT_FORWARD:   b.handlePodPortForward,

		aranyagopb.CMD_CRED_ENSURE: b.handleCredentialEnsure,
		aranyagopb.CMD_REJECT:      b.handleRejectCmd,

		aranyagopb.CMD_CTR_NET_ENSURE: b.handleContainerNetworkEnsure,
	}[cmd.Header.Kind]

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
