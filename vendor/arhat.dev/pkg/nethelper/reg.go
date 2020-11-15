// +build !nonethelper

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

package nethelper

import (
	"context"
	"net"

	"arhat.dev/pkg/wellknownerrors"
)

type key struct {
	network    string
	supportTLS bool
}

type bundle struct {
	listen ListenFunc
	dial   DialFunc
}

var (
	supportedNetwork = make(map[key]*bundle)
)

func Register(network string, supportTLS bool, listen ListenFunc, dial DialFunc) {
	supportedNetwork[key{
		network:    network,
		supportTLS: supportTLS,
	}] = &bundle{
		listen: listen,
		dial:   dial,
	}
}

func Dial(ctx context.Context, dialer interface{}, network, addr string, tlsConfig interface{}) (net.Conn, error) {
	n, ok := supportedNetwork[key{
		network:    network,
		supportTLS: tlsConfig != nil,
	}]
	if !ok {
		return nil, wellknownerrors.ErrNotSupported
	}

	return n.dial(ctx, dialer, network, addr, tlsConfig)
}

func Listen(
	ctx context.Context,
	config interface{},
	network, addr string,
	tlsConfig interface{},
) (interface{}, error) {
	n, ok := supportedNetwork[key{
		network:    network,
		supportTLS: tlsConfig != nil,
	}]
	if !ok {
		return nil, wellknownerrors.ErrNotSupported
	}

	return n.listen(ctx, config, network, addr, tlsConfig)
}
