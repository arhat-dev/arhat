// +build !nonethelper
// +build !nonethelper_stdnet

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

package stdnet

import (
	"context"
	"crypto/tls"
	"net"

	"arhat.dev/pkg/nethelper"
)

func init() {
	for _, n := range []string{"tcp", "tcp4", "tcp6", "unix"} {
		nethelper.Register(n, true, listenStream, dial)
		nethelper.Register(n, false, listenStream, dial)
	}

	for _, n := range []string{"udp", "udp4", "udp6", "unixpacket", "unixgram"} {
		nethelper.Register(n, false, listenPacket, dial)
	}
}

func dial(
	ctx context.Context,
	dialer interface{},
	network, addr string,
	tlsCfg interface{},
) (net.Conn, error) {
	var tlsConfig *tls.Config
	switch t := tlsCfg.(type) {
	case *tls.Config:
		tlsConfig = t
	case nil:
		tlsConfig = nil
	default:
		return nil, nethelper.ErrTLSConfigInvalid
	}

	var d *net.Dialer
	switch t := dialer.(type) {
	case *net.Dialer:
		d = t
	case nil:
		d = new(net.Dialer)
	default:
		return nil, nethelper.ErrDialerInvalid
	}

	conn, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	if tlsConfig == nil {
		return conn, nil
	}

	return tls.Client(conn, tlsConfig), nil
}

func listenStream(
	ctx context.Context,
	config interface{},
	network, addr string,
	tlsCfg interface{},
) (interface{}, error) {
	var tlsConfig *tls.Config
	switch t := tlsCfg.(type) {
	case *tls.Config:
		tlsConfig = t
	case nil:
		tlsConfig = nil
	default:
		return nil, nethelper.ErrTLSConfigInvalid
	}

	var lc *net.ListenConfig
	switch t := config.(type) {
	case *net.ListenConfig:
		lc = t
	case nil:
		lc = new(net.ListenConfig)
	default:
		return nil, nethelper.ErrListenConfigInvalid
	}

	l, err := lc.Listen(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	if tlsConfig == nil {
		return l, nil
	}

	return tls.NewListener(l, tlsConfig), nil
}

func listenPacket(
	ctx context.Context,
	config interface{},
	network, addr string,
	tlsConfig interface{},
) (interface{}, error) {
	_ = tlsConfig

	var lc *net.ListenConfig
	switch t := config.(type) {
	case *net.ListenConfig:
		lc = t
	case nil:
		lc = new(net.ListenConfig)
	default:
		return nil, nethelper.ErrListenConfigInvalid
	}

	return lc.ListenPacket(ctx, network, addr)
}
