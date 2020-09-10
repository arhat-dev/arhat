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

package runtimeutil

import (
	"context"
	"io"
	"net"
	"strconv"

	"arhat.dev/pkg/wellknownerrors"
)

// PortForward
// TODO: evaluate more efficient way to get network traffic redirected
func PortForward(ctx context.Context, address, protocol string, port int32, downstream io.ReadWriter) error {
	var (
		err    error
		dialer = &net.Dialer{}
	)

	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := dialer.DialContext(connCtx, protocol, net.JoinHostPort(address, strconv.FormatInt(int64(port), 10)))
	if err != nil {
		return err
	}

	go func() {
		<-connCtx.Done()
		_ = conn.Close()
	}()

	switch c := conn.(type) {
	case *net.TCPConn:
		err = handleTCPPortforward(c, downstream)
	case *net.UDPConn:
		err = handleUDPPortforward(c, downstream)
	default:
		return wellknownerrors.ErrNotSupported
	}

	if err != nil {
		return err
	}

	return nil
}

func handleTCPPortforward(upstream *net.TCPConn, downstream io.ReadWriter) error {
	// discard any unsent data
	_ = upstream.SetLinger(0)

	go func() {
		// take advantage of splice syscall
		if _, err := upstream.ReadFrom(downstream); checkNetError(err) != nil {
			// TODO: log error
			_ = err
		}
	}()

	if _, err := io.Copy(downstream, upstream); checkNetError(err) != nil {
		return err
	}

	return nil
}

func handleUDPPortforward(
	upstream *net.UDPConn,
	downstream io.ReadWriter,
) error {
	// https://github.com/kubernetes/kubernetes/issues/47862
	_ = upstream
	_ = downstream
	return wellknownerrors.ErrNotSupported
}

func checkNetError(err error) error {
	if err == nil {
		return nil
	}

	if netErr, ok := err.(*net.OpError); ok {
		switch netErr.Err {
		case io.EOF, io.ErrClosedPipe:
			// ignore these errors
			return nil
		}
	}

	return err
}
