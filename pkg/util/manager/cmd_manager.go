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

package manager

import (
	"runtime"
	"sync/atomic"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/queue"
)

func NewCmdManager() *CmdManager {
	return &CmdManager{
		sessionSQ:   make(map[uint64]*queue.SeqQueue),
		partialCmds: make(map[uint64]*[]byte),
	}
}

type CmdManager struct {
	sessionSQ   map[uint64]*queue.SeqQueue
	partialCmds map[uint64]*[]byte

	_working uint32
}

func (m *CmdManager) doExclusive(f func()) {
	for !atomic.CompareAndSwapUint32(&m._working, 0, 1) {
		runtime.Gosched()
	}

	f()

	atomic.StoreUint32(&m._working, 0)
}

func (m *CmdManager) Process(cmd *aranyagopb.Cmd) (cmdPayload []byte, complete bool) {
	// all in one cmd packet
	if cmd.Seq == 0 && cmd.Complete {
		return cmd.Payload, true
	}

	var (
		sid     = cmd.Sid
		sq      *queue.SeqQueue
		dataPtr *[]byte
	)
	m.doExclusive(func() {
		var ok bool
		dataPtr, ok = m.partialCmds[sid]
		if !ok {
			data := make([]byte, 0, 32)
			dataPtr = &data
			m.partialCmds[sid] = dataPtr
		}

		sq, ok = m.sessionSQ[sid]
		if !ok {
			sq = queue.NewSeqQueue(func(seq uint64, d interface{}) {
				*dataPtr = append(*dataPtr, d.([]byte)...)
			})
			m.sessionSQ[sid] = sq
		}
	})

	if cmd.Complete {
		complete = sq.SetMaxSeq(cmd.Seq)
	}

	complete = sq.Offer(cmd.Seq, cmd.Payload)
	if !complete {
		return
	}

	cmdPayload = *dataPtr

	delete(m.sessionSQ, sid)
	delete(m.partialCmds, sid)

	return
}
