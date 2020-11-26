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
	"io"
	"runtime"
	"sync"
	"sync/atomic"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/libext/extutil"
	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"
	"ext.arhat.dev/runtimeutil/networkutil"
	"ext.arhat.dev/runtimeutil/storageutil"
	"github.com/gogo/protobuf/proto"

	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/util/errconv"
	"arhat.dev/arhat/pkg/util/manager"
)

var (
	errClientNotSet            = errors.New("client not set")
	errRequiredOptionsNotFound = errors.New("required options not found")
	errCommandNotProvided      = errors.New("command not provided for exec")
)

type (
	rawCmdHandleFunc func(sid uint64, streamPreparing *uint32, data []byte)
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
			cmd, err2 := exechelper.Do(exechelper.Spec{
				Context: ctx,
				Command: config.Network.AbbotRequestExec,
				Env:     env,
				Stdin:   stdin,
				Stdout:  stdout,
				Stderr:  stderr,
			})
			if err2 != nil {
				return err2
			}

			_, err2 = cmd.Wait()
			return err2
		},
	)

	agent := &Agent{
		ctx:    appCtx,
		logger: logger,

		hostConfig:    &config.Arhat.Host,
		machineIDFrom: &config.Arhat.Node.MachineIDFrom,
		kubeLogFile:   config.Arhat.Log.KubeLogFile(),
		extInfo:       extInfo,

		cmdMgr:        manager.NewCmdManager(),
		storageClient: sc,
		networkClient: nc,

		streams: extutil.NewStreamManager(),

		pendingStreams: new(sync.Map),
	}

	err = agent.agentComponentExtension.init(agent, agent.logger, &config.Extension)
	if err != nil {
		return nil, fmt.Errorf("failed to init extension: %w", err)
	}

	err = agent.agentComponentPProf.init(agent.ctx, &config.Arhat.PProf)
	if err != nil {
		return nil, fmt.Errorf("failed to init pprof: %w", err)
	}

	err = agent.agentComponentMetrics.init()
	if err != nil {
		return nil, fmt.Errorf("failed to init metrics: %w", err)
	}

	agent.funcMap = map[aranyagopb.CmdType]rawCmdHandleFunc{
		aranyagopb.CMD_SESSION_CLOSE: agent.handleSessionClose,
		aranyagopb.CMD_REJECT:        agent.handleRejectCmd,

		aranyagopb.CMD_NET: agent.handleNetwork,

		// runtime cmd is handled differently, the payload will be sent to the
		// extension directly
		// aranyagopb.CMD_RUNTIME: nil,

		aranyagopb.CMD_NODE_INFO_GET: agent.handleNodeInfoGet,
		aranyagopb.CMD_EXEC:          agent.handleExec,
		aranyagopb.CMD_ATTACH:        agent.handleAttach,
		aranyagopb.CMD_LOGS:          agent.handleLogs,
		aranyagopb.CMD_TTY_RESIZE:    agent.handleTerminalResize,
		aranyagopb.CMD_PORT_FORWARD:  agent.handlePortForward,

		aranyagopb.CMD_METRICS_CONFIG:  agent.handleMetricsConfig,
		aranyagopb.CMD_METRICS_COLLECT: agent.handleMetricsCollect,

		aranyagopb.CMD_CRED_ENSURE: agent.handleCredentialEnsure,

		aranyagopb.CMD_PERIPHERAL_LIST:            agent.handlePeripheralList,
		aranyagopb.CMD_PERIPHERAL_ENSURE:          agent.handlePeripheralEnsure,
		aranyagopb.CMD_PERIPHERAL_DELETE:          agent.handlePeripheralDelete,
		aranyagopb.CMD_PERIPHERAL_OPERATE:         agent.handlePeripheralOperate,
		aranyagopb.CMD_PERIPHERAL_COLLECT_METRICS: agent.handlePeripheralMetricsCollect,
	}

	return agent, nil
}

// nolint:maligned
type Agent struct {
	ctx    context.Context
	logger log.Interface

	hostConfig    *conf.HostConfig
	machineIDFrom *conf.ValueFromSpec
	kubeLogFile   string
	extInfo       []*aranyagopb.NodeExtInfo

	cmdMgr *manager.CmdManager

	storageClient *storageutil.Client
	networkClient *networkutil.Client

	streams *extutil.StreamManager

	agentComponentPProf
	agentComponentMetrics
	agentComponentExtension

	settingClient uint32
	client        client.Interface

	funcMap map[aranyagopb.CmdType]rawCmdHandleFunc

	pendingStreams *sync.Map
}

func (b *Agent) SetClient(client client.Interface) {
	for !atomic.CompareAndSwapUint32(&b.settingClient, 0, 1) {
		runtime.Gosched()
	}
	b.client = client
	atomic.StoreUint32(&b.settingClient, 0)
}

func (b *Agent) GetClient() client.Interface {
	for !atomic.CompareAndSwapUint32(&b.settingClient, 0, 1) {
		runtime.Gosched()
	}
	c := b.client
	atomic.StoreUint32(&b.settingClient, 0)
	return c
}

func (b *Agent) PostData(sid uint64, kind aranyagopb.MsgType, seq uint64, completed bool, data []byte) (uint64, error) {
	c := b.GetClient()
	if c == nil {
		return seq, errClientNotSet
	}

	n := c.MaxPayloadSize()
	for len(data) > n {
		buf := data
		err := c.PostMsg(&aranyagopb.Msg{
			Kind:      kind,
			Sid:       sid,
			Seq:       seq,
			Completed: false,
			Payload:   buf[:n],
		})
		if err != nil {
			return seq, fmt.Errorf("failed to post msg chunk: %w", err)
		}

		seq++
		data = data[n:]
	}

	err := c.PostMsg(&aranyagopb.Msg{
		Kind:      kind,
		Sid:       sid,
		Seq:       seq,
		Completed: completed,
		Payload:   data,
	})
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

func (b *Agent) markStreamPrepared(sid uint64, preparing interface{}) {
	b.pendingStreams.Delete(sid)
	if preparing != nil {
		atomic.StoreUint32(preparing.(*uint32), 0)
	}
}

// HandleCmd may be called from any goroutine, no calling sequence guaranteed
func (b *Agent) HandleCmd(cmdBytes []byte) {
	if b.GetClient() == nil {
		b.handleConnectivityError(0, fmt.Errorf("client empty"))
		return
	}

	cmd := new(aranyagopb.Cmd)
	err := cmd.Unmarshal(cmdBytes)
	if err != nil {
		// discard invalid data
		b.logger.I("discard invalid cmd bytes", log.Binary("data", cmdBytes))
		return
	}

	sid := cmd.Sid

	// always assume is stream and mark stream as preparing
	streamPreparing := uint32(1)
	actual, loaded := b.pendingStreams.LoadOrStore(sid, &streamPreparing)

	// handle stream as special target, since data sequence is expected
	if cmd.Kind == aranyagopb.CMD_DATA_UPSTREAM {
		handleData := func() {
			if b.streams.Has(sid) {
				// is data for agent
				b.streams.Write(sid, cmd.Seq, cmd.Payload)
				if cmd.Completed {
					// write nil to mark max seq
					b.streams.Write(sid, cmd.Seq, nil)
				}
			} else {
				// is data for runtime
				err := b.sendRuntimeCmd(arhatgopb.CMD_DATA_INPUT, sid, cmd.Seq, cmd.Payload)
				if err != nil {
					b.handleRuntimeError(sid, err)
				}
			}
		}

		if !loaded {
			// was set by this cmd handle call, the stream is prepared
			b.markStreamPrepared(sid, &streamPreparing)

			handleData()
			return
		}

		// loaded preparing, wait until finished preparing
		perparing := actual.(*uint32)
		go func() {
			// wait for stream if not prepared
			for atomic.LoadUint32(perparing) != 0 {
				runtime.Gosched()
			}

			handleData()
		}()

		return
	}

	// reassamble cmd payload
	cmdPayload, complete := b.cmdMgr.Process(cmd)
	if !complete {
		b.logger.V("partial cmd not compelete")
		return
	}

	switch cmd.Kind {
	case aranyagopb.CMD_RUNTIME:
		b.markStreamPrepared(sid, &streamPreparing)

		// deliver runtime cmd
		err := b.sendRuntimeCmd(
			arhatgopb.CMD_RUNTIME_ARANYA_PROTO, cmd.Sid, 0, cmdPayload,
		)
		if err != nil {
			b.handleRuntimeError(sid, err)
		}
		return
	case aranyagopb.CMD_EXEC, aranyagopb.CMD_ATTACH, aranyagopb.CMD_PORT_FORWARD:
		// do not delete pending stream
	default:
		b.markStreamPrepared(sid, &streamPreparing)
	}

	handleCmd, ok := b.funcMap[cmd.Kind]
	if handleCmd == nil || !ok {
		b.handleUnknownCmd(sid, "unknown", cmd)
		return
	}

	handleCmd(sid, &streamPreparing, cmdPayload)
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
	b.logger.I(
		"cmd unhandled",
		log.String("kind", category),
		log.Uint64("sid", sid),
		log.Any("cmd", cmd),
	)

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
