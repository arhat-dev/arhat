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
package server

import (
	"runtime"
	"sync/atomic"

	"arhat.dev/arhat-proto/arhatgopb"
)

type msgWaitKey struct {
	id  uint64
	seq uint64
}

func newMsgWaitValue() *msgWaitValue {
	return &msgWaitValue{
		msgCh:   make(chan *arhatgopb.Msg, 1),
		closed:  0,
		working: 0,
	}
}

type msgWaitValue struct {
	msgCh chan *arhatgopb.Msg

	closed  uint32
	working uint32
}

func (m *msgWaitValue) do(f func()) {
	defer func() {
		for !atomic.CompareAndSwapUint32(&m.working, 1, 0) {
			runtime.Gosched()
		}
	}()
	for !atomic.CompareAndSwapUint32(&m.working, 0, 1) {
		runtime.Gosched()
	}

	f()
}

func (m *msgWaitValue) send(cancel <-chan struct{}, msg *arhatgopb.Msg) {
	m.do(func() {
		if atomic.LoadUint32(&m.closed) == 1 {
			return
		}
		select {
		case <-cancel:
		case m.msgCh <- msg:
		}
	})
}

func (m *msgWaitValue) close() {
	m.do(func() {
		if atomic.CompareAndSwapUint32(&m.closed, 0, 1) {
			close(m.msgCh)
		}
	})
}
