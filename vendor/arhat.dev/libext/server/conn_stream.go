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
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"

	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/libext/codec"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/nethelper"
)

type netConnectionHandleFunc func(
	listenAddr net.Addr,
	kind arhatgopb.ExtensionType,
	name string,
	codec codec.Interface,
	conn io.ReadWriter,
) error

type connectionManager interface {
	ListenAndServe() error
	Close() error
}

func newStreamConnectionManager(
	ctx context.Context,
	logger log.Interface,
	addr net.Addr,
	tlsConfig *tls.Config,
	handleFunc netConnectionHandleFunc,
) connectionManager {
	return &streamConnectionManager{
		baseConnectionManager: newBaseConnectionManager(ctx, logger, addr, handleFunc),
		tlsConfig:             tlsConfig,
	}
}

var _ connectionManager = (*streamConnectionManager)(nil)

type streamConnectionManager struct {
	*baseConnectionManager

	tlsConfig *tls.Config

	l io.Closer
}

func (m *streamConnectionManager) Close() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.l != nil {
		return m.l.Close()
	}

	return nil
}

func (m *streamConnectionManager) ListenAndServe() error {
	l, err := func() (net.Listener, error) {
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
		case net.Listener:
			m.l = t
			return t, nil
		case io.Closer:
			_ = t.Close()
		}

		return nil, fmt.Errorf("invalid %q network listener", m.addr.Network())
	}()
	if err != nil {
		return err
	}

	defer func() {
		_ = l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept new connection: %w", err)
		}

		m.logger.V("accepted new connection")
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
}
