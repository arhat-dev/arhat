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
)

var _ net.Addr = (*PipeAddr)(nil)

type PipeListener struct {
	net.Listener
}

type PipeConn struct {
	net.Conn
}

type PipeAddr struct {
	Path string
}

func (a *PipeAddr) Network() string {
	return "pipe"
}

func (a *PipeAddr) String() string {
	return a.Path
}

// ListenConfig for pipe listener, only intended for unix fifo
// nolint:maligned
type ListenConfig struct {
	// Options for unix-like system only

	// ConnectionDir for communication fifo
	ConnectionDir string

	// Permission of the listen fifo
	Permission uint32

	// Following options are for windows only

	// SecurityDescriptor contains a Windows security descriptor in SDDL format.
	SecurityDescriptor string

	// MessageMode determines whether the pipe is in byte or message mode. In either
	// case the pipe is read in byte mode by default. The only practical difference in
	// this implementation is that CloseWrite() is only supported for message mode pipes;
	// CloseWrite() is implemented as a zero-byte write, but zero-byte writes are only
	// transferred to the reader (and returned as io.EOF in this implementation)
	// when the pipe is in message mode.
	MessageMode bool

	// InputBufferSize specifies the size the input buffer, in bytes.
	InputBufferSize uint64

	// OutputBufferSize specifies the size the input buffer, in bytes.
	OutputBufferSize uint64
}

func ListenPipe(path string) (net.Listener, error) {
	return (&ListenConfig{}).ListenPipe(path)
}

// Dialer for pipe network
type Dialer struct {
	LocalPath string
}

func Dial(path string) (net.Conn, error) {
	return DialPipe(nil, &PipeAddr{Path: path})
}

func DialContext(ctx context.Context, path string) (net.Conn, error) {
	return DialPipeContext(ctx, nil, &PipeAddr{Path: path})
}

func DialPipe(laddr *PipeAddr, raddr *PipeAddr) (net.Conn, error) {
	return DialPipeContext(context.Background(), laddr, raddr)
}

func DialPipeContext(ctx context.Context, laddr *PipeAddr, raddr *PipeAddr) (_ net.Conn, err error) {
	var localPath string
	if laddr != nil {
		localPath = laddr.Path
	}

	return (&Dialer{LocalPath: localPath}).DialPipeContext(ctx, raddr)
}
