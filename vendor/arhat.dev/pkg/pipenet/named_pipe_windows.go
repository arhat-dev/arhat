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
	"net"

	"github.com/Microsoft/go-winio"
)

func (c *ListenConfig) ListenPipe(path string) (net.Listener, error) {
	l, err := winio.ListenPipe(path, &winio.PipeConfig{
		SecurityDescriptor: c.SecurityDescriptor,
		MessageMode:        c.MessageMode,
		InputBufferSize:    int32(c.InputBufferSize),
		OutputBufferSize:   int32(c.OutputBufferSize),
	})
	if err != nil {
		return nil, err
	}

	return &PipeListener{l}, nil
}

func (d *Dialer) Dial(path string) (net.Conn, error) {
	return d.DialPipe(&PipeAddr{Path: path})
}

func (d *Dialer) DialContext(ctx context.Context, path string) (net.Conn, error) {
	return d.DialPipeContext(ctx, &PipeAddr{Path: path})
}

func (d *Dialer) DialPipe(raddr *PipeAddr) (net.Conn, error) {
	return d.DialPipeContext(context.Background(), raddr)
}

func (d *Dialer) DialPipeContext(ctx context.Context, raddr *PipeAddr) (net.Conn, error) {
	conn, err := winio.DialPipeContext(ctx, raddr.Path)
	if err != nil {
		return nil, err
	}

	return &PipeConn{conn}, nil
}
