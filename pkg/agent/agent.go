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

package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"

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
	errNilCmd                  = errors.New("cmd is nil")
	errRequiredOptionsNotFound = errors.New("required options not found")
	errStreamSessionClosed     = errors.New("stream session closed")
	errCommandNotProvided      = errors.New("command not provided for exec")
)

var _ types.Agent = &Agent{}

func NewAgent(appCtx context.Context, config *conf.ArhatConfig) (*Agent, error) {
	ctx, exit := context.WithCancel(appCtx)

	extInfo, err := convertNodeExtInfo(config.Arhat.Node.ExtInfo)
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

		clientStore: new(atomic.Value),
		runtime:     rt,
		storage:     st,
		devices:     newDeviceManager(),
		streams:     manager.NewStreamManager(),
	}
	agent.streams.Start(ctx.Done())

	return agent, nil
}

type Agent struct {
	hostConfig    *conf.ArhatHostConfig
	maxPodAvail   uint64
	machineIDFrom *conf.ArhatValueFromSpec
	kubeLogFile   string
	extInfo       []*aranyagopb.NodeExtInfo

	ctx  context.Context
	exit context.CancelFunc

	logger log.Interface

	metricsMU               *sync.RWMutex
	collectNodeMetrics      types.MetricsCollectFunc
	collectContainerMetrics types.MetricsCollectFunc

	clientStore *atomic.Value
	runtime     types.Runtime
	storage     types.Storage
	devices     *deviceManager
	streams     *manager.StreamManager
}

func (b *Agent) SetClient(client types.AgentConnectivity) {
	b.clientStore.Store(client)
}

func (b *Agent) GetClient() types.AgentConnectivity {
	if c := b.clientStore.Load(); c != nil {
		return c.(types.AgentConnectivity)
	}

	return nil
}

func (b *Agent) PostMsg(msg *aranyagopb.Msg) error {
	if c := b.clientStore.Load(); c != nil {
		return c.(types.AgentConnectivity).PostMsg(msg)
	}

	return errClientNotSet
}

func (b *Agent) Context() context.Context {
	return b.ctx
}

func (b *Agent) HandleCmd(cmd *aranyagopb.Cmd) {
	if b.GetClient() == nil {
		b.handleConnectivityError(0, fmt.Errorf("client empty"))
		return
	}

	if len(cmd.Body) == 0 {
		b.handleRuntimeError(cmd.SessionId, errNilCmd)
		return
	}

	switch cmd.Kind {
	case aranyagopb.CMD_NODE:
		b.handleNodeCmd(cmd.SessionId, cmd.Body)
	case aranyagopb.CMD_DEVICE:
		b.handleDeviceCmd(cmd.SessionId, cmd.Body)
	case aranyagopb.CMD_METRICS:
		b.handleMetricsCmd(cmd.SessionId, cmd.Body)
	case aranyagopb.CMD_POD:
		b.handlePodCmd(cmd.SessionId, cmd.Body)
	case aranyagopb.CMD_POD_OPERATION:
		b.handlePodOperationCmd(cmd.SessionId, cmd.Body)
	case aranyagopb.CMD_CRED:
		b.handleCredentialCmd(cmd.SessionId, cmd.Body)
	case aranyagopb.CMD_REJECTION:
		b.handleRejectCmd(cmd.SessionId, cmd.Body)
	case aranyagopb.CMD_SESSION:
		b.handleSessionCmd(cmd.SessionId, cmd.Body)
	case aranyagopb.CMD_NETWORK:
		b.handleNetworkCmd(cmd.SessionId, cmd.Body)
	default:
		b.handleUnknownCmd(cmd.SessionId, "unknown", cmd)
	}
}

func (b *Agent) processInNewGoroutine(sid uint64, cmdName string, process func()) {
	go func() {
		b.logger.V("working on", log.Uint64("sid", sid), log.String("work", cmdName))
		process()
		b.logger.V("finished", log.Uint64("sid", sid), log.String("work", cmdName))
	}()
}

func (b *Agent) handleSyncLoop(sid uint64, name string, opt *aranyagopb.SyncOptions, doSync func()) {
	if opt == nil {
		b.handleRuntimeError(sid, errRequiredOptionsNotFound)
		return
	}

	client := b.GetClient()
	if client == nil {
		b.handleRuntimeError(sid, errClientNotSet)
		return
	}

	clientExit := client.Context().Done()
	agentExit := b.ctx.Done()
	interval := time.Duration(opt.Interval)

	b.processInNewGoroutine(sid, name+".sync.loop", func() {
		// sync until connection lost
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				// send global sync message
				doSync()
			case <-clientExit:
				return
			case <-agentExit:
				return
			}
		}
	})
}

func (b *Agent) handleUnknownCmd(sid uint64, category string, cmd interface{}) bool {
	b.logger.I(fmt.Sprintf("unknown %s cmd", category), log.Uint64("sid", sid), log.Any("cmd", cmd))
	return b.handleRuntimeError(sid, wellknownerrors.ErrNotSupported)
}

func (b *Agent) handleRuntimeError(sid uint64, err error) bool {
	if err == nil {
		return false
	}

	plainErr := b.PostMsg(aranyagopb.NewErrorMsg(sid, errconv.ToConnectivityError(err)))
	if plainErr != nil {
		b.handleConnectivityError(sid, plainErr)
	}

	return true
}

func (b *Agent) handleConnectivityError(sid uint64, err error) bool {
	if err == nil {
		return false
	}

	b.logger.I("connectivity error", log.Uint64("sid", sid), log.Error(err))
	return true
}
