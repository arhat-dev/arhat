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
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/queue"
)

func NewCmdManager() *CmdManager {
	return &CmdManager{
		sessionSQ:   make(map[uint64]*queue.SeqQueue),
		partialCmds: make(map[uint64][]byte),
		mu:          new(sync.RWMutex),
	}
}

type CmdManager struct {
	sessionSQ   map[uint64]*queue.SeqQueue
	partialCmds map[uint64][]byte

	mu *sync.RWMutex
}

func (m *CmdManager) Process(cmd *aranyagopb.Cmd) (cmdPayload []byte, complete bool) {
	// all in one cmd packet
	if cmd.Seq == 0 && cmd.Completed {
		return cmd.Payload, true
	}

	sid := cmd.Sid

	m.mu.Lock()
	defer m.mu.Unlock()

	sq, ok := m.sessionSQ[sid]
	if !ok {
		sq = queue.NewSeqQueue()
		m.sessionSQ[sid] = sq
	}

	cmdByteChunks, complete := sq.Offer(cmd.Seq, cmd.Payload)
	for _, ck := range cmdByteChunks {
		if ck == nil {
			continue
		}

		m.partialCmds[sid] = append(m.partialCmds[sid], ck.([]byte)...)
	}

	if cmd.Completed {
		complete = sq.SetMaxSeq(cmd.Seq)
	}

	if !complete {
		return
	}

	cmdPayload = m.partialCmds[sid]

	delete(m.sessionSQ, sid)
	delete(m.partialCmds, sid)

	return
}
