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

package extutil

import (
	"io"
	"runtime"
	"sync/atomic"

	"arhat.dev/pkg/queue"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/libext/types"
)

func noopHandleResize(cols, rows uint32) {}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		sessions: make(map[uint64]*stream),
	}
}

type StreamManager struct {
	sessions map[uint64]*stream

	_working uint32
}

func (m *StreamManager) doExclusive(f func()) {
	for !atomic.CompareAndSwapUint32(&m._working, 0, 1) {
		runtime.Gosched()
	}

	f()

	atomic.StoreUint32(&m._working, 0)
}

func (m *StreamManager) Has(sid uint64) (exists bool) {
	m.doExclusive(func() {
		_, exists = m.sessions[sid]
	})

	return
}

func (m *StreamManager) Add(sid uint64, create func() (io.WriteCloser, types.ResizeHandleFunc, error)) (err error) {
	m.doExclusive(func() {
		if _, ok := m.sessions[sid]; ok {
			err = wellknownerrors.ErrAlreadyExists
			return
		}

		var (
			w  io.WriteCloser
			rF types.ResizeHandleFunc
		)
		w, rF, err = create()
		if err != nil {
			return
		}

		if rF == nil {
			rF = noopHandleResize
		}

		m.sessions[sid] = &stream{
			_w:  w,
			_rF: rF,

			_seqQ: queue.NewSeqQueue(),
		}
	})

	return
}

func (m *StreamManager) Del(sid uint64) {
	m.doExclusive(func() {
		if s, ok := m.sessions[sid]; ok {
			s.close()
			delete(m.sessions, sid)
		}
	})
}

func (m *StreamManager) Write(sid, seq uint64, data []byte) bool {
	var (
		s  *stream
		ok bool
	)

	m.doExclusive(func() {
		s, ok = m.sessions[sid]
	})
	if !ok {
		return false
	}

	s.write(seq, data)
	return true
}

func (m *StreamManager) Resize(sid uint64, cols, rows uint32) {
	var (
		s  *stream
		ok bool
	)
	m.doExclusive(func() {
		s, ok = m.sessions[sid]
	})
	if !ok {
		return
	}

	s.resize(cols, rows)
}

type stream struct {
	_seqQ *queue.SeqQueue

	_w  io.WriteCloser
	_rF types.ResizeHandleFunc

	_working uint32
}

func (m *stream) doExclusive(f func()) {
	for !atomic.CompareAndSwapUint32(&m._working, 0, 1) {
		runtime.Gosched()
	}

	f()

	atomic.StoreUint32(&m._working, 0)
}

func (s *stream) write(seq uint64, data []byte) {
	s.doExclusive(func() {
		if len(data) == 0 {
			_ = s._seqQ.SetMaxSeq(seq)
		}

		out, complete := s._seqQ.Offer(seq, data)
		for _, d := range out {
			_, _ = s._w.Write(d.([]byte))
		}

		if complete {
			_ = s._w.Close()
		}
	})
}

func (s *stream) resize(cols, rows uint32) {
	s._rF(cols, rows)
}

func (s *stream) close() {
	s.doExclusive(func() {
		_ = s._w.Close()
	})
}
