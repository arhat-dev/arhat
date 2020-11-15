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
	"io"
)

const (
	defaultPacketReadBufSize = 65537
)

// Forward network traffic
// the parameters:
// 	ctx is used to cancel dial operation
// 	dialer is optional for custom network dial options
// 	network is the network name, e.g. tcp, udp, tcp4
// 	addr is the endpoint address
// 	upstream is the data channel to the endpoint
// 	packetReadBuf is the buffer used for udp/ip/unix connection
//
// the return values:
// 	downstream is used to read data sent from the forwarded port and close connection
// 	closeWrite is intended to close write in stream oriented connection
// 	readErrCh is used to check read error and whether donwstream reading finished
//	err if not nil the port forward failed
func Forward(
	ctx context.Context,
	dialer interface{},
	network string,
	addr string,
	upstream io.Reader,
	packetReadBuf []byte,
) (
	downstream io.ReadCloser,
	closeWrite func(),
	readErrCh <-chan error,
	err error,
) {
	conn, err := Dial(ctx, dialer, network, addr, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	// find close write support
	switch c := conn.(type) {
	case interface {
		CloseWrite() error
	}:
		closeWrite = func() {
			_ = c.CloseWrite()
		}
	default:
		// nop
		closeWrite = func() {}
	}

	// find read from support
	errCh := make(chan error)
	switch c := conn.(type) {
	case io.ReaderFrom:
		// take advantage of splice syscall if possible (usually used in tcp)
		go func() {
			defer close(errCh)

			_, err2 := c.ReadFrom(upstream)
			if err2 != nil {
				select {
				case errCh <- err2:
				case <-ctx.Done():
				}
			}
		}()
	default:
		// other kind of connections will use copy
		go func() {
			defer close(errCh)

			if len(packetReadBuf) == 0 {
				packetReadBuf = make([]byte, defaultPacketReadBufSize)
			}

			_, err2 := io.CopyBuffer(conn, upstream, packetReadBuf)
			if err2 != nil {
				select {
				case <-ctx.Done():
				case errCh <- err2:
				}
			}
		}()
	}

	return conn, closeWrite, errCh, nil
}
