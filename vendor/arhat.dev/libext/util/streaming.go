package util

import (
	"io"
	"runtime"
	"sync"
	"sync/atomic"

	"arhat.dev/pkg/queue"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/libext/types"
)

func noopHandleResize(cols, rows uint32) {}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		sessions: make(map[uint64]*stream),
		mu:       new(sync.RWMutex),
	}
}

type StreamManager struct {
	sessions map[uint64]*stream

	mu *sync.RWMutex
}

func (m *StreamManager) Add(sid uint64, create func() (io.WriteCloser, types.ResizeHandleFunc, error)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[sid]; ok {
		return wellknownerrors.ErrAlreadyExists
	}

	w, rF, err := create()
	if err != nil {
		return err
	}

	if rF == nil {
		rF = noopHandleResize
	}

	m.sessions[sid] = &stream{
		_w:  w,
		_rF: rF,

		_seqQ: queue.NewSeqQueue(),
	}

	return nil
}

func (m *StreamManager) Del(sid uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[sid]; ok {
		s.close()
		delete(m.sessions, sid)
	}
}

func (m *StreamManager) Write(sid, seq uint64, data []byte) {
	m.mu.RLock()
	s, ok := m.sessions[sid]
	m.mu.RUnlock()
	if !ok {
		return
	}

	s.write(seq, data)
}

func (m *StreamManager) Resize(sid uint64, cols, rows uint32) {
	m.mu.RLock()
	s, ok := m.sessions[sid]
	m.mu.RUnlock()

	if !ok {
		return
	}

	s.resize(cols, rows)
}

type stream struct {
	_seqQ *queue.SeqQueue

	_w  io.WriteCloser
	_rF types.ResizeHandleFunc

	working uint32
}

func (s *stream) write(seq uint64, data []byte) {
	for !atomic.CompareAndSwapUint32(&s.working, 0, 1) {
		runtime.Gosched()
	}

	if data == nil {
		s._seqQ.SetMaxSeq(seq)
		return
	}

	out, _ := s._seqQ.Offer(seq, data)
	for _, d := range out {
		_, _ = s._w.Write(d.([]byte))
	}

	atomic.StoreUint32(&s.working, 0)
}

func (s *stream) resize(cols, rows uint32) {
	s._rF(cols, rows)
}

func (s *stream) close() {
	for !atomic.CompareAndSwapUint32(&s.working, 0, 1) {
		runtime.Gosched()
	}

	_ = s._w.Close()

	atomic.StoreUint32(&s.working, 0)
}
