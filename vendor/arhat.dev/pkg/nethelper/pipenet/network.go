// +build !nonethelper
// +build !nonethelper_pipenet

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

package pipenet

import (
	"context"
	"crypto/tls"
	"net"

	"arhat.dev/pkg/nethelper"
	"arhat.dev/pkg/pipenet"
)

func init() {
	nethelper.Register("pipe", false, listen, dial)
	nethelper.Register("pipe", true, listen, dial)
}

func dial(
	ctx context.Context,
	dialer interface{},
	network, addr string,
	tlsCfg interface{},
) (net.Conn, error) {
	_ = network

	var tlsConfig *tls.Config
	switch t := tlsCfg.(type) {
	case *tls.Config:
		tlsConfig = t
	case nil:
		tlsConfig = nil
	default:
		return nil, nethelper.ErrTLSConfigInvalid
	}

	var d *pipenet.Dialer
	switch t := dialer.(type) {
	case *pipenet.Dialer:
		d = t
	case nil:
		d = new(pipenet.Dialer)
	default:
		return nil, nethelper.ErrDialerInvalid
	}

	conn, err := d.DialContext(ctx, addr)
	if err != nil {
		return nil, err
	}

	if tlsConfig == nil {
		return conn, nil
	}

	return tls.Client(conn, tlsConfig), nil
}

func listen(
	ctx context.Context,
	config interface{},
	network, addr string,
	tlsCfg interface{},
) (interface{}, error) {
	_ = ctx
	_ = network

	var tlsConfig *tls.Config
	switch t := tlsCfg.(type) {
	case *tls.Config:
		tlsConfig = t
	case nil:
		tlsConfig = nil
	default:
		return nil, nethelper.ErrTLSConfigInvalid
	}

	var lc *pipenet.ListenConfig
	switch t := config.(type) {
	case *pipenet.ListenConfig:
		lc = t
	case nil:
		lc = new(pipenet.ListenConfig)
	default:
		return nil, nethelper.ErrListenConfigInvalid
	}

	l, err := lc.ListenPipe(addr)
	if err != nil {
		return nil, err
	}

	if tlsConfig == nil {
		return l, nil
	}

	return tls.NewListener(l, tlsConfig), nil
}
