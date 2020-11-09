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

func (m *CmdManager) Process(cmd *aranyagopb.Cmd) (cmdBytes []byte, complete bool) {
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

	cmdBytes = m.partialCmds[sid]

	delete(m.sessionSQ, sid)
	delete(m.partialCmds, sid)

	return
}
