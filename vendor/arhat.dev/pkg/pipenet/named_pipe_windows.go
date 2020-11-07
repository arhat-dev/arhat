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
	"os"

	"github.com/Microsoft/go-winio"
)

func ListenPipe(path, connDir string, perm os.FileMode) (net.Listener, error) {
	_, _ = perm, connDir
	return winio.ListenPipe(path, &winio.PipeConfig{
		SecurityDescriptor: "",
		MessageMode:        false,
		InputBufferSize:    0,
		OutputBufferSize:   0,
	})
}

func Dial(path string) (net.Conn, error) {
	return DialPipe(nil, &PipeAddr{Path: path})
}

func DialContext(ctx context.Context, path string) (net.Conn, error) {
	return DialPipeContext(ctx, nil, &PipeAddr{Path: path})
}

func DialPipe(laddr *PipeAddr, raddr *PipeAddr) (net.Conn, error) {
	return DialPipeContext(context.TODO(), laddr, raddr)
}

func DialPipeContext(ctx context.Context, laddr *PipeAddr, raddr *PipeAddr) (net.Conn, error) {
	_ = laddr
	return winio.DialPipeContext(ctx, raddr.Path)
}
