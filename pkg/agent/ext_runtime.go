// +build !noextension
// +build !noextension_runtime

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
	"runtime"
	"sync/atomic"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/types"
	"arhat.dev/libext/server"
	"arhat.dev/pkg/wellknownerrors"
)

type extensionComponentRuntime struct {
	workingOnConn uintptr
	runtimeCtx    *server.ExtensionContext

	postData types.AgentDataPostFunc

	workingOnPost uintptr
	sessionSeq    map[uint64]uint64
}

func (c *extensionComponentRuntime) init(
	agent *Agent,
	srv *server.Server,
	config *conf.RuntimeExtensionConfig,
) {
	c.postData = agent.PostData
	srv.Handle(arhatgopb.EXTENSION_RUNTIME, func(extensionName string) (
		server.ExtensionHandleFunc, server.OutOfBandMsgHandleFunc,
	) {
		return c.handleRuntimeConn, c.handleRuntimeMsg
	})
}

func (c *extensionComponentRuntime) handleRuntimeConn(ctx *server.ExtensionContext) {
	for !atomic.CompareAndSwapUintptr(&c.workingOnConn, 0, 1) {
		runtime.Gosched()
	}

	ok := c.runtimeCtx == nil
	if ok {
		c.runtimeCtx = ctx
	}

	for !atomic.CompareAndSwapUintptr(&c.workingOnConn, 1, 0) {
		runtime.Gosched()
	}

	if !ok {
		// runtime already connected, do not accept new
		return
	}

	// wait until conneciton lost
	<-ctx.Context.Done()

	for !atomic.CompareAndSwapUintptr(&c.workingOnConn, 0, 1) {
		runtime.Gosched()
	}

	c.runtimeCtx = nil

	for !atomic.CompareAndSwapUintptr(&c.workingOnConn, 1, 0) {
		runtime.Gosched()
	}
}

func (c *extensionComponentRuntime) handleRuntimeMsg(msg *arhatgopb.Msg) {
	var kind aranyagopb.MsgType
	switch msg.Kind {
	case arhatgopb.MSG_DATA_OUTPUT:
		kind = aranyagopb.MSG_DATA
	case arhatgopb.MSG_RUNTIME_DATA_STDERR:
		kind = aranyagopb.MSG_DATA_STDERR
	case arhatgopb.MSG_RUNTIME_ARANYA_PROTO:
		_, _ = c.postData(msg.Id, aranyagopb.MSG_RUNTIME, 0, true, msg.Payload)
	default:
		// invalid runtime message data, discard
		return
	}

	for !atomic.CompareAndSwapUintptr(&c.workingOnPost, 0, 1) {
		runtime.Gosched()
	}

	sid := msg.Id
	seq, ok := c.sessionSeq[sid]
	if !ok {
		seq = msg.Ack
	}

	for !atomic.CompareAndSwapUintptr(&c.workingOnPost, 1, 0) {
		runtime.Gosched()
	}

	complete := len(msg.Payload) == 0
	lastSeq, err := c.postData(sid, kind, seq, complete, msg.Payload)

	for !atomic.CompareAndSwapUintptr(&c.workingOnPost, 0, 1) {
		runtime.Gosched()
	}

	if complete {
		delete(c.sessionSeq, sid)
	} else {
		c.sessionSeq[sid] = lastSeq
	}

	for !atomic.CompareAndSwapUintptr(&c.workingOnPost, 1, 0) {
		runtime.Gosched()
	}

	if err != nil {
		return
	}
}

func (c *extensionComponentRuntime) sendRuntimeCmd(kind arhatgopb.CmdType, sid, seq uint64, data []byte) error {
	for !atomic.CompareAndSwapUintptr(&c.workingOnConn, 0, 1) {
		runtime.Gosched()
	}

	ctx := c.runtimeCtx

	for !atomic.CompareAndSwapUintptr(&c.workingOnConn, 1, 0) {
		runtime.Gosched()
	}

	if ctx == nil {
		return wellknownerrors.ErrNotSupported
	}

	_, err := ctx.SendCmd(&arhatgopb.Cmd{
		Kind:    kind,
		Id:      sid,
		Seq:     seq,
		Payload: data,
	}, false)

	return err
}
