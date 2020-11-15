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
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"

	"arhat.dev/pkg/iohelper"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/nethelper"

	"arhat.dev/libext/codec"
)

func newPacketStream(p net.PacketConn, ra net.Addr, r io.ReadCloser) *packetStream {
	return &packetStream{
		conn:   p,
		ra:     ra,
		r:      bufio.NewReader(r),
		rClose: r,
	}
}

type packetStream struct {
	conn   net.PacketConn
	ra     net.Addr
	r      *bufio.Reader
	rClose io.Closer
}

func (p *packetStream) Read(data []byte) (n int, err error) {
	return p.r.Read(data)
}

func (p *packetStream) Write(data []byte) (n int, err error) {
	return p.conn.WriteTo(data, p.ra)
}

func (p *packetStream) Close() error {
	return p.rClose.Close()
}

func newPacketConnectionManager(
	ctx context.Context,
	logger log.Interface,
	addr net.Addr,
	tlsConfig *tls.Config,
	handleFunc netConnectionHandleFunc,
) connectionManager {
	return &packetConnectionManager{
		baseConnectionManager: newBaseConnectionManager(ctx, logger, addr, handleFunc),
		tlsConfig:             tlsConfig,
		connections:           new(sync.Map),
	}
}

var _ connectionManager = (*packetConnectionManager)(nil)

type packetConnectionManager struct {
	*baseConnectionManager

	l io.Closer

	tlsConfig   *tls.Config
	connections *sync.Map
}

func (m *packetConnectionManager) Close() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.l != nil {
		return m.l.Close()
	}

	return nil
}

func (m *packetConnectionManager) ListenAndServe() error {
	p, err := func() (io.Closer, error) {
		m.mu.Lock()
		defer m.mu.Unlock()

		var (
			lRaw interface{}
			err  error
		)
		if m.tlsConfig == nil {
			lRaw, err = nethelper.Listen(m.ctx, nil, m.addr.Network(), m.addr.String(), nil)
		} else {
			lRaw, err = nethelper.Listen(m.ctx, nil, m.addr.Network(), m.addr.String(), m.tlsConfig)
		}
		if err != nil {
			return nil, err
		}

		switch t := lRaw.(type) {
		case io.Closer:
			m.l = t
			return m.l, nil
		default:
			return nil, fmt.Errorf("invalid %q listener, no close interface", m.addr.Network())
		}
	}()
	if err != nil {
		return err
	}

	defer func() {
		_ = p.Close()
	}()

	switch t := p.(type) {
	case net.Listener:
		for {
			conn, err := t.Accept()
			if err != nil {
				return fmt.Errorf("failed to accept new packet connection: %w", err)
			}

			m.logger.V("accepted new packet connection")
			go func() {
				defer func() {
					_ = conn.Close()
				}()

				kind, name, codec, err2 := m.validateConnection(conn)
				if err2 != nil {
					m.logger.I("connection invalid", log.Error(err2))
					return
				}

				err2 = m.handleNewConn(m.addr, kind, name, codec, conn)
				if err2 != nil {
					m.logger.I("failed to handle new connection", log.Error(err2))
					return
				}
			}()
		}
	case net.PacketConn:
		buf := make([]byte, 65535)
		for {
			n, ra, err2 := t.ReadFrom(buf)
			if err2 != nil {
				return fmt.Errorf("failed to read packet: %w", err2)
			}

			data := codec.GetBytesBuf(n)
			_ = copy(data[:n], buf[:n])

			addr := ra.String()
			v, ok := m.connections.Load(addr)
			if ok {
				pw := v.(io.WriteCloser)
				_, err2 = pw.Write(data[:n])
				if err2 != nil {
					_ = pw.Close()

					m.logger.I("failed to write data to packet connection")
				}

				codec.PutBytesBuf(&data)
				continue
			}

			// not existing connection for this address, create a new one
			kind, name, c, err2 := m.validateConnection(bytes.NewReader(data[:n]))
			codec.PutBytesBuf(&data)
			if err2 != nil {
				m.logger.I("invalid new packet connection", log.Error(err2))
				continue
			}

			pr, pw := iohelper.Pipe()
			conn := newPacketStream(t, ra, pr)

			m.connections.Store(addr, pw)
			go func() {
				defer func() {
					_ = pw.Close()
					m.connections.Delete(addr)
				}()

				err3 := m.handleNewConn(m.addr, kind, name, c, conn)
				if err3 != nil {
					m.logger.I("failed to handle new packet connection", log.Error(err3))
					return
				}
			}()
		}
	default:
		return fmt.Errorf("unexpected conn")
	}
}
