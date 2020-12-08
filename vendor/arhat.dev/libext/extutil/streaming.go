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
	"math"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

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

		m.sessions[sid] = newStream(w, rF)
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

func newStream(w io.WriteCloser, rF types.ResizeHandleFunc) *stream {
	s := &stream{
		_w:  w,
		_rF: rF,

		closed:    make(chan struct{}),
		seqDataCh: make(chan []byte, 16),

		next: 0,
		max:  math.MaxUint64,

		dataChunks:      make([]*seqData, 0, 16),
		currentChunkPtr: nil,
	}

	atomic.StorePointer(&s.currentChunkPtr, unsafe.Pointer(&s.dataChunks))
	go s.handleWrite()

	return s
}

type stream struct {
	_w  io.WriteCloser
	_rF types.ResizeHandleFunc

	closed    chan struct{}
	seqDataCh chan []byte

	// embedded sequence queue
	next uint64
	max  uint64

	dataChunks      []*seqData
	currentChunkPtr unsafe.Pointer
	_working        uint32
}

func (s *stream) handleWrite() {
	for data := range s.seqDataCh {
		_, _ = s._w.Write(data)
	}
}

func (s *stream) write(seq uint64, data []byte) {
	if len(data) == 0 {
		_ = s.setMaxSeq(seq)
	}

	if s.offer(seq, data) {
		s.close()
	}
}

func (s *stream) resize(cols, rows uint32) {
	s._rF(cols, rows)
}

func (s *stream) close() {
	select {
	case <-s.closed:
		return
	default:
		close(s.closed)
	}

	_ = s._w.Close()

	// wait until no data remain in data channel
	for {
		// close is issued after offer(), so we just need to check
		// if there is data remaining in the buffered channel
		if len(s.seqDataCh) == 0 {
			close(s.seqDataCh)
			return
		}

		time.Sleep(time.Second)
	}
}

type seqData struct {
	seq  uint64
	data []byte
}

func (s *stream) offer(seq uint64, data []byte) (complete bool) {
	// take a snapshot of existing values
	var (
		currentNext = atomic.LoadUint64(&s.next)
		currentMax  = atomic.LoadUint64(&s.max)
	)

	switch {
	case currentNext > currentMax:
		// already complete, discard
		return true
	case seq > currentMax, seq < currentNext:
		// exceeded or duplicated, discard
		return false
	case seq == currentNext:
		currentNext++

		if !atomic.CompareAndSwapUint64(&s.next, currentNext-1, currentNext) {
			// dup, discard
			return currentNext > currentMax
		}

		// is expected next chunk, pop it and its following chunks

		select {
		case <-s.closed:
			return
		case s.seqDataCh <- data:
		}

		currentChunks := *(*[]*seqData)(atomic.LoadPointer(&s.currentChunkPtr))
		if len(currentChunks) == 0 {
			return currentNext > currentMax
		}

		trimStart := -1
		for i, chunk := range currentChunks {
			if chunk.seq != currentNext || chunk.seq > currentMax {
				break
			}

			if !atomic.CompareAndSwapUint64(&s.next, currentNext, currentNext+1) {
				// handled in another groutine, should not happen, just in case
				return
			}

			select {
			case <-s.closed:
			case s.seqDataCh <- chunk.data:
			}

			currentNext++

			trimStart = i
		}

		s.doExclusive(func() {
			// we only take consactive chunks from the start
			// so it's safe to use index directly
			s.dataChunks = s.dataChunks[trimStart+1:]
			atomic.StorePointer(&s.currentChunkPtr, unsafe.Pointer(&s.dataChunks))
		})

		return currentNext > currentMax
	}

	// cache unordered data chunk

	currentChunks := *(*[]*seqData)(atomic.LoadPointer(&s.currentChunkPtr))
	insertAt := 0
	for i, chunk := range currentChunks {
		if chunk.seq > seq {
			insertAt = i
			break
		}

		// duplicated
		if chunk.seq == seq {
			return false
		}

		insertAt = i + 1
	}

	s.doExclusive(func() {
		s.dataChunks = append(
			s.dataChunks[:insertAt],
			append(
				[]*seqData{{seq: seq, data: data}},
				s.dataChunks[insertAt:]...,
			)...,
		)

		atomic.StorePointer(&s.currentChunkPtr, unsafe.Pointer(&s.dataChunks))
	})

	return false
}

func (s *stream) setMaxSeq(maxSeq uint64) (completed bool) {
	if currentNext := atomic.LoadUint64(&s.next); currentNext > maxSeq {
		// existing seq data already exceeds maxSeq
		atomic.StoreUint64(&s.max, currentNext)
		return true
	}

	atomic.StoreUint64(&s.max, maxSeq)
	return false
}

func (s *stream) doExclusive(f func()) {
	for !atomic.CompareAndSwapUint32(&s._working, 0, 1) {
		runtime.Gosched()
	}

	f()

	atomic.StoreUint32(&s._working, 0)
}
