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
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/pipenet"
	"golang.org/x/sync/errgroup"

	"arhat.dev/libext/codec"
)

type Config struct {
	Endpoints []EndpointConfig
}

type EndpointConfig struct {
	Listen            string
	TLS               *tls.Config
	KeepaliveInterval time.Duration
	MessageTimeout    time.Duration
}

func NewServer(
	ctx context.Context,
	logger log.Interface,
	config *Config,
) (*Server, error) {
	if len(config.Endpoints) == 0 {
		return nil, fmt.Errorf("must provide at least one listen url")
	}

	srv := &Server{
		ctx:               ctx,
		logger:            logger,
		connectionConfig:  make(map[string]*EndpointConfig),
		extensionManagers: make(map[extensionKey]*ExtensionManager),
		mu:                new(sync.RWMutex),
	}

	for i, ep := range config.Endpoints {
		u, err := url.Parse(ep.Listen)
		if err != nil {
			return nil, fmt.Errorf("invalid listen url: %w", err)
		}

		var (
			addr          net.Addr
			createConnMgr func(
				context.Context,
				log.Interface,
				net.Addr,
				*tls.Config,
				netConnectionHandleFunc,
			) connectionManager
		)
		switch s := strings.ToLower(u.Scheme); s {
		case "tcp", "tcp4", "tcp6": // nolint:goconst
			addr, err = net.ResolveTCPAddr(s, u.Host)
			createConnMgr = newStreamConnectionManager
		case "udp", "udp4", "udp6": // nolint:goconst
			addr, err = net.ResolveUDPAddr(s, u.Host)
			createConnMgr = newPacketConnectionManager
		case "unix": // nolint:goconst
			addr, err = net.ResolveUnixAddr(s, u.Path)
			createConnMgr = newStreamConnectionManager
		case "pipe":
			err = nil
			switch runtime.GOOS {
			case "windows":
				// pipe://PipeName
				addr = &pipenet.PipeAddr{Path: fmt.Sprintf(`\\.\pipe\%s%s`, u.Host, u.Path)}
			default:
				addr = &pipenet.PipeAddr{Path: u.Path}
			}
			createConnMgr = newStreamConnectionManager
		default:
			return nil, fmt.Errorf("unsupported protocol %q", u.Scheme)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s listen address: %w", u.Scheme, err)
		}
		connMgr := createConnMgr(
			ctx, logger, addr, ep.TLS, srv.handleNewConn,
		)

		srv.connectionConfig[addr.Network()+"://"+addr.String()] = &config.Endpoints[i]

		srv.connectionManagers = append(srv.connectionManagers, connMgr)
	}

	return srv, nil
}

type extensionKey struct {
	kind arhatgopb.ExtensionType
	name string
}

type Server struct {
	ctx    context.Context
	logger log.Interface

	connectionConfig   map[string]*EndpointConfig
	connectionManagers []connectionManager
	extensionManagers  map[extensionKey]*ExtensionManager

	mu *sync.RWMutex
}

// ListenAndServe listen on local addresses and accept connections from extensions
func (s *Server) ListenAndServe() error {
	wg, ctx := errgroup.WithContext(s.ctx)
	_ = ctx
	for _, m := range s.connectionManagers {
		mgr := m
		wg.Go(mgr.ListenAndServe)
	}

	err := wg.Wait()

	for _, m := range s.connectionManagers {
		_ = m.Close()
	}

	return err
}

// Close all listening endpoints
func (s *Server) Close() {
	for _, m := range s.connectionManagers {
		_ = m.Close()
	}
}

// Handle extension session with handleFunc, you can specify optional extension names (`extensions`) to restrict
// the handleFunc to these extensions, if no extension name specified, this handleFunc will override all
// existing handleFunc for extensions with this kind
func (s *Server) Handle(kind arhatgopb.ExtensionType, handleFactory ExtensionHandleFuncFactory, extensions ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mgr := NewExtensionManager(s.ctx, s.logger, handleFactory)
	for i, name := range extensions {
		if name == "" {
			return
		}
		s.extensionManagers[extensionKey{
			kind: kind,
			name: extensions[i],
		}] = mgr
	}

	if len(extensions) == 0 {
		// handle all kinds of extensions in this kind
		var toRemove []extensionKey
		for k := range s.extensionManagers {
			if k.kind != kind {
				continue
			}

			toRemove = append(toRemove, extensionKey{
				kind: kind,
				name: k.name,
			})
		}

		for _, k := range toRemove {
			delete(s.extensionManagers, k)
		}

		s.extensionManagers[extensionKey{
			kind: kind,
			name: "",
		}] = mgr
	}
}

func (s *Server) handleNewConn(
	listenAddr net.Addr,
	kind arhatgopb.ExtensionType,
	name string,
	codec codec.Interface,
	conn io.ReadWriter,
) error {
	s.mu.RLock()
	mgr, ok := s.extensionManagers[extensionKey{
		kind: kind,
		name: name,
	}]
	s.mu.RUnlock()

	if !ok {
		// fallback to general extension
		s.mu.RLock()
		mgr, ok = s.extensionManagers[extensionKey{
			kind: kind,
			name: "",
		}]
		s.mu.RUnlock()
	}

	if !ok {
		return fmt.Errorf("no handler for %s", kind.String())
	}

	ec, ok := s.connectionConfig[listenAddr.Network()+"://"+listenAddr.String()]
	if !ok {
		return fmt.Errorf("endpoint config not found")
	}

	keepaliveInterval := ec.KeepaliveInterval
	messageTimeout := ec.MessageTimeout

	if keepaliveInterval == 0 {
		keepaliveInterval = time.Minute
	}

	if keepaliveInterval < time.Millisecond {
		keepaliveInterval = time.Millisecond
	}

	if messageTimeout == 0 {
		messageTimeout = time.Minute
	}

	if messageTimeout < time.Millisecond {
		messageTimeout = time.Millisecond
	}

	return mgr.HandleStream(name, codec, keepaliveInterval, messageTimeout, conn)
}
